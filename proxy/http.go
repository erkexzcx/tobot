package proxyfier

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

func checkTarget(transport http.RoundTripper) error {
	originalTransport := httpClient.Transport
	defer func() {
		httpClient.Transport = originalTransport
	}()

	httpClient.Transport = transport

	req, err := http.NewRequest("GET", proxiesCheckTarget, nil)
	if err != nil {
		return errors.New("failed to create request to check target: " + err.Error())
	}
	req.Header.Set("User-Agent", proxiesCheckUserAgent)
	req.Header.Set("Host", proxiesCheckTargetHost)
	req.Header.Set("Accept", "*/*")

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.New("failed to send request to check target: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		return errors.New("unexpected response code: " + resp.Status)
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("failed to read response body: " + err.Error())
	}

	if strings.Contains(string(contents), "Sistema nustatė, jog jūs jungiates per kitą serverį, todėl greičiausiai bandote naudotis autokėlėju.") {
		return errors.New("IP ban detected")
	}

	return nil
}

func getTransportWithProxy(proxyString string) (transport http.RoundTripper, err error) {
	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		return nil, err
	}

	switch proxyURL.Scheme {
	case "http":
		return &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}, nil
	case "socks4":
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, errors.New("failed to create SOCKS4 dialer: " + err.Error())
		}
		return &http.Transport{
			Dial: dialer.Dial,
		}, nil
	case "socks5":
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
		if err != nil {
			return nil, errors.New("failed to create SOCKS5 dialer: " + err.Error())
		}
		return &http.Transport{
			Dial: dialer.Dial,
		}, nil
	default:
		return nil, errors.New("unsupported proxy scheme: " + proxyURL.Scheme)
	}
}
