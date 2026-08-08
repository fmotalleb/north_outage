package telegram

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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
		Transport: pollingAwareTransport{base: transport},
		Timeout:   15 * time.Second,
	}
	return client
}

// pollingAwareTransport instruments every outbound Telegram API request with
// OTel spans except the long-polling getUpdates calls. getUpdates runs
// continuously in the background and is expected to time out under network
// hiccups, so its spans (and their error statuses) are excluded from traces.
type pollingAwareTransport struct {
	base http.RoundTripper
}

func (t pollingAwareTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if isGetUpdatesRequest(req) {
		return t.base.RoundTrip(req)
	}
	return otelhttp.NewTransport(t.base).RoundTrip(req)
}

func isGetUpdatesRequest(req *http.Request) bool {
	return strings.HasSuffix(req.URL.Path, "/getUpdates")
}
