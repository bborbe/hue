// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

type Proxy func(req *http.Request) (*url.URL, error)

type CheckRedirect func(req *http.Request, via []*http.Request) error

type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

type HttpClientBuilder interface {
	WithRetry(retryLimit int, retryDelay time.Duration) HttpClientBuilder
	WithoutRetry() HttpClientBuilder
	WithProxy() HttpClientBuilder
	WithoutProxy() HttpClientBuilder
	// WithRedirects controls how many redirects are allowed
	// 0 = no redirects, -1 = infinit redirects, 10 = 10 max redirects
	WithRedirects(maxRedirect int) HttpClientBuilder
	// WithoutRedirects is equal to WithRedirects(0)
	WithoutRedirects() HttpClientBuilder
	WithTimeout(timeout time.Duration) HttpClientBuilder
	WithDialFunc(dialFunc DialFunc) HttpClientBuilder
	WithInsecureSkipVerify(insecureSkipVerify bool) HttpClientBuilder
	WithClientCert(caCertPath string, clientCertPath string, clientKeyPath string) HttpClientBuilder
	Build(ctx context.Context) (*http.Client, error)
	BuildRoundTripper(ctx context.Context) (http.RoundTripper, error)
}

func NewClientBuilder() HttpClientBuilder {
	b := new(httpClientBuilder)
	b.WithoutProxy()
	b.WithRedirects(10)
	b.WithTimeout(30 * time.Second)
	b.WithoutRetry()
	return b
}

type httpClientBuilder struct {
	proxy Proxy
	// maxRedirect -1 = infinit, 0 = none, and other number limits the redirects
	maxRedirect        int
	timeout            time.Duration
	dialFunc           DialFunc
	insecureSkipVerify bool
	retryLimit         int
	retryDelay         time.Duration
	caCertPath         string
	clientCertPath     string
	clientKeyPath      string
}

func (h *httpClientBuilder) WithRetry(retryLimit int, retryDelay time.Duration) HttpClientBuilder {
	h.retryLimit = retryLimit
	h.retryDelay = retryDelay
	return h
}

func (h *httpClientBuilder) WithoutRetry() HttpClientBuilder {
	h.retryLimit = 0
	h.retryDelay = 0
	return h
}

func (h *httpClientBuilder) WithClientCert(caCertPath string, clientCertPath string, clientKeyPath string) HttpClientBuilder {
	h.caCertPath = caCertPath
	h.clientCertPath = clientCertPath
	h.clientKeyPath = clientKeyPath
	return h
}

func (h *httpClientBuilder) WithTimeout(timeout time.Duration) HttpClientBuilder {
	h.timeout = timeout
	return h
}

func (h *httpClientBuilder) WithDialFunc(dialFunc DialFunc) HttpClientBuilder {
	h.dialFunc = dialFunc
	return h
}

func (h *httpClientBuilder) BuildDialFunc() DialFunc {
	if h.dialFunc != nil {
		return h.dialFunc
	}
	return (&net.Dialer{
		Timeout: h.timeout,
	}).DialContext
}

func (h *httpClientBuilder) BuildRoundTripper(ctx context.Context) (http.RoundTripper, error) {
	if glog.V(5) {
		glog.Infof("build http roundTripper")
	}
	tlsClientConfig, err := h.createTlsConfig(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create tlsConfig failed")
	}
	var roundTripper http.RoundTripper = &http.Transport{
		Proxy:           h.proxy,
		DialContext:     h.BuildDialFunc(),
		TLSClientConfig: tlsClientConfig,
	}
	if h.retryDelay > 0 && h.retryLimit > 0 {
		roundTripper = NewRoundTripperRetry(roundTripper, h.retryLimit, h.retryDelay)
	}
	return roundTripper, nil
}

func (h *httpClientBuilder) createTlsConfig(ctx context.Context) (*tls.Config, error) {
	tlsClientConfig := &tls.Config{}

	if h.caCertPath != "" && h.clientCertPath != "" && h.clientKeyPath != "" {
		var err error
		tlsClientConfig, err = CreateTlsClientConfig(ctx, h.caCertPath, h.clientCertPath, h.clientKeyPath)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "create tls config failed")
		}
	}
	tlsClientConfig.InsecureSkipVerify = h.insecureSkipVerify

	return tlsClientConfig, nil
}

func (h *httpClientBuilder) Build(ctx context.Context) (*http.Client, error) {
	if glog.V(5) {
		glog.Infof("build http client")
	}
	roundTripper, err := h.BuildRoundTripper(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build roundTripper failed")
	}

	return &http.Client{
		Transport:     roundTripper,
		CheckRedirect: h.createCheckRedirect(),
	}, nil
}

func (h *httpClientBuilder) WithProxy() HttpClientBuilder {
	h.proxy = http.ProxyFromEnvironment
	return h
}

func (h *httpClientBuilder) WithoutProxy() HttpClientBuilder {
	h.proxy = nil
	return h
}

func (h *httpClientBuilder) WithRedirects(maxRedirect int) HttpClientBuilder {
	h.maxRedirect = maxRedirect
	return h
}

func (h *httpClientBuilder) WithoutRedirects() HttpClientBuilder {
	h.maxRedirect = 0
	return h
}

func (h *httpClientBuilder) WithInsecureSkipVerify(insecureSkipVerify bool) HttpClientBuilder {
	h.insecureSkipVerify = insecureSkipVerify
	return nil
}

func (h *httpClientBuilder) createCheckRedirect() func(req *http.Request, via []*http.Request) error {
	switch h.maxRedirect {
	case -1:
		return func(req *http.Request, via []*http.Request) error {
			return nil
		}
	case 0:
		return func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	default:
		return func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return stderrors.New("stopped after 10 redirects")
			}
			return nil
		}
	}
}
