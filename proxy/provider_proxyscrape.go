package proxyfier

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type ProxyscrapeJSON struct {
	Proxies []struct {
		Proxy string `json:"proxy"`
	} `json:"proxies"`
}

func refresherProxyscrape() {
	for {
		proxiesMux.Lock()
		httpClient.Transport = http.DefaultTransport
		resp, err := httpClient.Get("https://api.proxyscrape.com/v4/free-proxy-list/get?request=display_proxies&proxy_format=protocolipport&format=json")
		if err != nil {
			log.Println("Failed to get proxy list from proxyscrape:", err)
			proxiesMux.Unlock()
			time.Sleep(15 * time.Minute)
			continue
		}

		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body from proxyscrape:", err)
			proxiesMux.Unlock()
			time.Sleep(15 * time.Minute)
			continue
		}

		var proxyscrape ProxyscrapeJSON
		err = json.Unmarshal(contents, &proxyscrape)
		if err != nil {
			log.Println("Failed to unmarshal proxyscrape response:", err)
			proxiesMux.Unlock()
			time.Sleep(15 * time.Minute)
			continue
		}

		proxiesList := make([]string, 0, len(proxyscrape.Proxies))
		for _, p := range proxyscrape.Proxies {
			proxiesList = append(proxiesList, p.Proxy)
		}

		for _, p := range proxiesList {
			proxies[p] = struct{}{}
		}

		proxiesMux.Unlock()
		time.Sleep(1 * time.Hour)
	}
}
