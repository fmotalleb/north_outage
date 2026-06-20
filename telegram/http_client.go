package telegram

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

func httpClient(proxyURL *url.URL) *http.Client {
	dialer := proxy.FromEnvironmentUsing(proxy.Direct)

	if proxyURL != nil {
		d, err := proxy.FromURL(proxyURL, dialer)
		if err != nil {
			panic("failed to use given url as proxy")
		}
		dialer = d
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
	return client
}
