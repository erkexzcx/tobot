package proxyfier

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	_ "github.com/bdandy/go-socks4"
)

var (
	dynamicProviders = []string{"proxyscrape"}

	proxiesEnabled         = false
	proxiesCheckTarget     = ""
	proxiesCheckTargetHost = ""
	proxiesCheckUserAgent  = ""

	proxies    = map[string]struct{}{}
	proxiesMux = sync.Mutex{}

	workingProxy          = ""
	workingProxyTransport = http.DefaultTransport
	workingProxyMux       = sync.Mutex{}
)

func Start(enabled bool, providers []string, static []string, checkTarget string, checkUserAgent string) error {
	workingProxyMux.Lock()
	defer workingProxyMux.Unlock()

	proxiesMux.Lock()
	defer proxiesMux.Unlock()

	if !enabled {
		return nil
	}
	proxiesEnabled = true
	proxiesCheckTarget = checkTarget
	proxiesCheckUserAgent = checkUserAgent
	targetURL, _ := url.Parse(checkTarget)
	proxiesCheckTargetHost = targetURL.Host

	for _, p := range static {
		proxies[p] = struct{}{}
	}

	for _, p := range providers {
		found := false
		for _, dp := range dynamicProviders {
			if p == dp {
				found = true
				break
			}
		}
		if !found {
			return errors.New("unknown dynamic proxy provider: " + p)
		}

		switch p {
		case "proxyscrape":
			go refresherProxyscrape()
		}
	}

	go startRefresher()

	return nil
}

func GetTransport() http.RoundTripper {
	workingProxyMux.Lock()

	if !proxiesEnabled {
		workingProxyMux.Unlock()
		return workingProxyTransport
	}

	if workingProxy == "" {
		workingProxyMux.Unlock()
		log.Println("No working proxy available, retrying...")
		time.Sleep(1 * time.Second)
		return GetTransport()
	}

	defer workingProxyMux.Unlock()
	return workingProxyTransport
}

func startRefresher() {
	for {

		// Get list of proxies
		proxiesMux.Lock()
		proxiesCopy := make([]string, 0, len(proxies))
		for p := range proxies {
			proxiesCopy = append(proxiesCopy, p)
		}
		proxiesMux.Unlock()
		if len(proxiesCopy) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		// Get working proxy
		proxiesMux.Lock()
		workingProxyCopy := workingProxy

		if workingProxyCopy != "" {
			t, err := getTransportWithProxy(workingProxyCopy)
			if err != nil {
				log.Println("(REFRESHER) Failed to create transport with working proxy "+workingProxyCopy+":", err)
			} else {
				if err := checkTarget(t); err != nil {
					log.Println("(REFRESHER) Proxy " + workingProxyCopy + " is not working: " + err.Error())
					delete(proxies, workingProxyCopy)
				} else {
					// Same working proxy is still working
					proxiesMux.Unlock()
					time.Sleep(10 * time.Second)
					continue
				}
			}
		}

		// If we reached this point - need new proxy
		for _, p := range proxiesCopy {
			t, err := getTransportWithProxy(p)
			if err != nil {
				log.Println("(REFRESHER) Failed to create transport with proxy "+p+":", err)
				continue
			}
			if err := checkTarget(t); err != nil {
				log.Println("(REFRESHER) Proxy " + p + " is not working " + err.Error())
				delete(proxies, p)
			} else {
				workingProxy = p
				workingProxyTransport = t
				log.Println("(REFRESHER) Working proxy set to " + p)
				break
			}
		}

		proxiesMux.Unlock()
	}
}
