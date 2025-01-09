package req

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/imroc/req/v3"
)

func UpdateTransport(tr *req.Transport, proxyUrl *url.URL) {
	tr.TLSClientConfig.InsecureSkipVerify = true
	tr.TLSClientConfig.Renegotiation = tls.RenegotiateOnceAsClient
	tr.DisableKeepAlives = true
	tr.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext
	tr.MaxIdleConns = 100
	tr.IdleConnTimeout = 90 * time.Second
	tr.TLSHandshakeTimeout = 10 * time.Second
	tr.ExpectContinueTimeout = 1 * time.Second
	tr.MaxIdleConnsPerHost = 100
	tr.MaxResponseHeaderBytes = 4096 // net/http default is 10Mb
	if proxyUrl != nil {
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
}

func NewHTTPClient() *req.Client {
	client := req.NewClient()
	client.ImpersonateChrome()
	client.SetTimeout(120 * time.Second)
	return client
}

func NewDefaultHTTPClient() *req.Client {
	client := NewHTTPClient()
	UpdateTransport(client.GetTransport(), nil)
	return client
}
