package player

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

func (p *Player) httpRequest(method, link string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, link, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", HEADER_HOST)
	req.Header.Set("User-Agent", HEADER_USER_AGENT)
	req.Header.Set("Accept", "*/*")

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		linkURL, err := url.Parse(link)
		if err != nil {
			return nil, errors.New("unable to parse link " + link)
		}
		redirectURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			return nil, errors.New("unable to parse HTTP header \"Location\" of link " + link + " after redirection")
		}
		newLink := linkURL.ResolveReference(redirectURL)
		return p.httpRequest(method, newLink.String(), body)
	}

	return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
}
