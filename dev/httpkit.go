package httpkit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

// Default connection pool settings.
const (
	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10
	DefaultIdleConnTimeout     = 90 * time.Second
)

// Client is the package-level HTTP client.
var Client *HTTPClient

// HTTPClient wraps http.Client with optional concurrency-safe updates.
type HTTPClient struct {
	client                   *http.Client
	baseTransport            *http.Transport
	defaultProxy             *ProxyConfig
	defaultTimeout           time.Duration
	defaultTLSConfig         *tls.Config
	defaultJar               http.CookieJar
	defaultDisableKeepAlives bool
	defaultFollowRedirects   *bool
	defaultMaxRedirects      int
	proxyClients             map[string]*http.Client
	mu                       sync.RWMutex
}

// ProxyConfig defines proxy settings.
type ProxyConfig struct {
	Type     string `json:"type"`    // "http" or "socks5"
	Network  string `json:"network"` // "tcp" or "udp" (socks5 only)
	Address  string `json:"address"` // e.g. "127.0.0.1:8080"
	Username string `json:"username"`
	Password string `json:"password"`
}

// RequestOptions controls per-request behavior.
type RequestOptions struct {
	Headers           map[string]string `json:"headers"`
	Timeout           time.Duration     `json:"timeout"`
	Proxy             *ProxyConfig      `json:"proxy"`
	Context           context.Context   `json:"-"` // Not serialized
	TLSConfig         *tls.Config       `json:"-"`
	Jar               http.CookieJar    `json:"-"`
	DisableKeepAlives *bool             `json:"disable_keep_alives"`
	FollowRedirects   *bool             `json:"follow_redirects"`
	MaxRedirects      int               `json:"max_redirects"`
	RetryCount        int               `json:"retry_count"`
	RetryDelay        time.Duration     `json:"retry_delay"`
	MaxResponseSize   int64             `json:"max_response_size"` // Max body size in bytes
}

// Response is the result of an HTTP request.
type Response struct {
	StatusCode int            `json:"status_code"`
	Headers    http.Header    `json:"headers"`
	Body       []byte         `json:"body"`
	RawResp    *http.Response `json:"-"`        // Raw response, not serialized
	Duration   time.Duration  `json:"duration"` // Total duration
}

func init() {
	Client = NewHTTPClient(nil)
}

// NewStdClient returns a configured *http.Client using RequestOptions.
// This is useful when callers need a standard client (e.g., for SDKs).
func NewStdClient(options *RequestOptions) *http.Client {
	hc := NewHTTPClient(options)
	if options == nil {
		return hc.getClient()
	}
	return hc.getClientForOptions(options)
}

// NewHTTPClient creates a new HTTPClient with default pooling settings.
func NewHTTPClient(options *RequestOptions) *HTTPClient {
	defaultTimeout := 30 * time.Second
	if options != nil && options.Timeout > 0 {
		defaultTimeout = options.Timeout
	}

	defaultTLSConfig := (*tls.Config)(nil)
	if options != nil && options.TLSConfig != nil {
		defaultTLSConfig = options.TLSConfig
	}

	defaultJar := http.CookieJar(nil)
	if options != nil && options.Jar != nil {
		defaultJar = options.Jar
	}

	defaultDisableKeepAlives := false
	if options != nil && options.DisableKeepAlives != nil {
		defaultDisableKeepAlives = *options.DisableKeepAlives
	}

	var defaultFollowRedirects *bool
	if options != nil && options.FollowRedirects != nil {
		defaultFollowRedirects = options.FollowRedirects
	}

	defaultMaxRedirects := 0
	if options != nil && options.MaxRedirects > 0 {
		defaultMaxRedirects = options.MaxRedirects
	}

	baseTransport := &http.Transport{
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   defaultDisableKeepAlives,
		TLSClientConfig:     defaultTLSConfig,
	}

	var defaultProxy *ProxyConfig
	clientTransport := baseTransport
	if options != nil && options.Proxy != nil {
		defaultProxy = options.Proxy
		proxyTransport := baseTransport.Clone()
		if err := configureProxy(proxyTransport, options.Proxy); err == nil {
			clientTransport = proxyTransport
		} else {
			defaultProxy = nil
		}
	}

	client := &http.Client{
		Transport:     clientTransport,
		Jar:           defaultJar,
		CheckRedirect: buildRedirectPolicy(defaultFollowRedirects, defaultMaxRedirects),
	}

	defaultKey := clientKey(defaultProxy, defaultTLSConfig, defaultDisableKeepAlives, defaultJar, defaultFollowRedirects, defaultMaxRedirects)
	proxyClients := map[string]*http.Client{defaultKey: client}

	return &HTTPClient{
		client:                   client,
		baseTransport:            baseTransport,
		defaultProxy:             defaultProxy,
		defaultTimeout:           defaultTimeout,
		defaultTLSConfig:         defaultTLSConfig,
		defaultJar:               defaultJar,
		defaultDisableKeepAlives: defaultDisableKeepAlives,
		defaultFollowRedirects:   defaultFollowRedirects,
		defaultMaxRedirects:      defaultMaxRedirects,
		proxyClients:             proxyClients,
		mu:                       sync.RWMutex{},
	}
}

func (c *HTTPClient) getClient() *http.Client {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	return client
}

func (c *HTTPClient) getDefaultTimeout() time.Duration {
	c.mu.RLock()
	timeout := c.defaultTimeout
	c.mu.RUnlock()
	return timeout
}

func proxyKey(proxyConfig *ProxyConfig) string {
	if proxyConfig == nil {
		return ""
	}
	return strings.Join([]string{
		strings.ToLower(proxyConfig.Type),
		strings.ToLower(proxyConfig.Network),
		proxyConfig.Address,
		proxyConfig.Username,
		proxyConfig.Password,
	}, "|")
}

func boolPtrKey(value *bool) string {
	if value == nil {
		return "nil"
	}
	if *value {
		return "true"
	}
	return "false"
}

func clientKey(proxyConfig *ProxyConfig, tlsConfig *tls.Config, disableKeepAlives bool, jar http.CookieJar, followRedirects *bool, maxRedirects int) string {
	return strings.Join([]string{
		proxyKey(proxyConfig),
		fmt.Sprintf("%p", tlsConfig),
		fmt.Sprintf("%p", jar),
		strconv.FormatBool(disableKeepAlives),
		boolPtrKey(followRedirects),
		strconv.Itoa(maxRedirects),
	}, "|")
}

func buildRedirectPolicy(followRedirects *bool, maxRedirects int) func(req *http.Request, via []*http.Request) error {
	if followRedirects != nil && !*followRedirects {
		return func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	if maxRedirects > 0 {
		return func(req *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return nil
		}
	}
	return nil
}

func (c *HTTPClient) getClientForOptions(opt *RequestOptions) *http.Client {
	c.mu.RLock()
	defaultProxy := c.defaultProxy
	defaultTLSConfig := c.defaultTLSConfig
	defaultJar := c.defaultJar
	defaultDisableKeepAlives := c.defaultDisableKeepAlives
	defaultFollowRedirects := c.defaultFollowRedirects
	defaultMaxRedirects := c.defaultMaxRedirects
	baseTransport := c.baseTransport
	c.mu.RUnlock()

	effectiveProxy := defaultProxy
	if opt != nil && opt.Proxy != nil {
		effectiveProxy = opt.Proxy
	}

	effectiveTLSConfig := defaultTLSConfig
	if opt != nil && opt.TLSConfig != nil {
		effectiveTLSConfig = opt.TLSConfig
	}

	effectiveJar := defaultJar
	if opt != nil && opt.Jar != nil {
		effectiveJar = opt.Jar
	}

	effectiveDisableKeepAlives := defaultDisableKeepAlives
	if opt != nil && opt.DisableKeepAlives != nil {
		effectiveDisableKeepAlives = *opt.DisableKeepAlives
	}

	effectiveFollowRedirects := defaultFollowRedirects
	if opt != nil && opt.FollowRedirects != nil {
		effectiveFollowRedirects = opt.FollowRedirects
	}

	effectiveMaxRedirects := defaultMaxRedirects
	if opt != nil && opt.MaxRedirects > 0 {
		effectiveMaxRedirects = opt.MaxRedirects
	}

	key := clientKey(effectiveProxy, effectiveTLSConfig, effectiveDisableKeepAlives, effectiveJar, effectiveFollowRedirects, effectiveMaxRedirects)

	c.mu.RLock()
	if c.proxyClients != nil {
		if client, ok := c.proxyClients[key]; ok {
			c.mu.RUnlock()
			return client
		}
	}
	c.mu.RUnlock()

	if baseTransport == nil {
		return c.getClient()
	}

	transport := baseTransport.Clone()
	transport.DisableKeepAlives = effectiveDisableKeepAlives
	transport.TLSClientConfig = effectiveTLSConfig
	if effectiveProxy != nil {
		if err := configureProxy(transport, effectiveProxy); err != nil {
			return c.getClient()
		}
	}

	client := &http.Client{
		Transport:     transport,
		Jar:           effectiveJar,
		CheckRedirect: buildRedirectPolicy(effectiveFollowRedirects, effectiveMaxRedirects),
	}

	c.mu.Lock()
	if c.proxyClients == nil {
		c.proxyClients = make(map[string]*http.Client)
	}
	if existing, ok := c.proxyClients[key]; ok {
		c.mu.Unlock()
		return existing
	}
	c.proxyClients[key] = client
	c.mu.Unlock()

	return client
}

// configureProxy applies proxy settings to the transport.
func configureProxy(transport *http.Transport, proxyConfig *ProxyConfig) error {
	switch strings.ToLower(proxyConfig.Type) {
	case "http", "https":
		scheme := strings.ToLower(proxyConfig.Type)
		proxyURL := fmt.Sprintf("%s://%s", scheme, proxyConfig.Address)
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			proxyURL = fmt.Sprintf("%s://%s:%s@%s",
				scheme,
				url.QueryEscape(proxyConfig.Username),
				url.QueryEscape(proxyConfig.Password),
				proxyConfig.Address)
		}

		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedURL)

	case "socks5":
		var auth *proxy.Auth
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			auth = &proxy.Auth{
				User:     proxyConfig.Username,
				Password: proxyConfig.Password,
			}
		}

		network := "tcp"
		if proxyConfig.Network != "" {
			network = proxyConfig.Network
		}

		dialer, err := proxy.SOCKS5(network, proxyConfig.Address, auth, proxy.Direct)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
			transport.DialContext = contextDialer.DialContext
		} else {
			return fmt.Errorf("SOCKS5 dialer does not support context")
		}
	default:
		return fmt.Errorf("unsupported proxy type: %s", proxyConfig.Type)
	}

	return nil
}

// Get sends a GET request.
func (c *HTTPClient) Get(url string, options ...*RequestOptions) (*Response, error) {
	return c.Request("GET", url, nil, options...)
}

// Post sends a POST request.
func (c *HTTPClient) Post(url string, body interface{}, options ...*RequestOptions) (*Response, error) {
	return c.Request("POST", url, body, options...)
}

// Put sends a PUT request.
func (c *HTTPClient) Put(url string, body interface{}, options ...*RequestOptions) (*Response, error) {
	return c.Request("PUT", url, body, options...)
}

// Delete sends a DELETE request.
func (c *HTTPClient) Delete(url string, options ...*RequestOptions) (*Response, error) {
	return c.Request("DELETE", url, nil, options...)
}

type bodyProvider func() (io.Reader, string, error)

// buildBodyProvider creates a reusable body reader for retries.
func buildBodyProvider(body interface{}, retryCount int) (bodyProvider, error) {
	if body == nil {
		return nil, nil
	}

	switch v := body.(type) {
	case string:
		data := v
		return func() (io.Reader, string, error) {
			return strings.NewReader(data), "text/plain", nil
		}, nil
	case []byte:
		data := make([]byte, len(v))
		copy(data, v)
		return func() (io.Reader, string, error) {
			return bytes.NewReader(data), "application/octet-stream", nil
		}, nil
	case io.ReadSeeker:
		return func() (io.Reader, string, error) {
			if _, err := v.Seek(0, io.SeekStart); err != nil {
				return nil, "", err
			}
			return v, "", nil
		}, nil
	case io.Reader:
		if retryCount > 1 {
			data, err := io.ReadAll(v)
			if err != nil {
				return nil, err
			}
			return func() (io.Reader, string, error) {
				return bytes.NewReader(data), "", nil
			}, nil
		}
		return func() (io.Reader, string, error) {
			return v, "", nil
		}, nil
	default:
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		return func() (io.Reader, string, error) {
			return bytes.NewReader(jsonData), "application/json", nil
		}, nil
	}
}

// isIdempotentMethod reports whether a method is safe to retry.
func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "PUT", "DELETE", "OPTIONS", "TRACE":
		return true
	default:
		return false
	}
}

// Request sends an HTTP request with optional retries and context.
func (c *HTTPClient) Request(method, url string, body interface{}, options ...*RequestOptions) (*Response, error) {
	var opt *RequestOptions
	if len(options) > 0 && options[0] != nil {
		opt = options[0]
	} else {
		opt = &RequestOptions{}
	}

	ctx := opt.Context
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := c.getDefaultTimeout()
	if opt.Timeout > 0 {
		timeout = opt.Timeout
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	retryCount := opt.RetryCount
	if retryCount <= 0 {
		retryCount = 1
	}
	if !isIdempotentMethod(method) && retryCount > 1 {
		retryCount = 1
	}

	bodyFn, err := buildBodyProvider(body, retryCount)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare body: %w", err)
	}

	var lastErr error
	for i := 0; i < retryCount; i++ {
		if i > 0 && opt.RetryDelay > 0 {
			select {
			case <-time.After(opt.RetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := c.doSingleRequest(ctx, method, url, bodyFn, opt)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if !shouldRetry(err) {
			break
		}
	}

	return nil, lastErr
}

// doSingleRequest executes a single request attempt.
func (c *HTTPClient) doSingleRequest(ctx context.Context, method, url string, bodyFn bodyProvider, opt *RequestOptions) (*Response, error) {
	start := time.Now()

	var bodyReader io.Reader
	var contentType string
	var err error
	if bodyFn != nil {
		bodyReader, contentType, err = bodyFn()
		if err != nil {
			return nil, fmt.Errorf("failed to prepare body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if bodyFn != nil && req.Header.Get("Content-Type") == "" && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, value := range opt.Headers {
		req.Header.Set(key, value)
	}

	client := c.getClientForOptions(opt)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if opt.MaxResponseSize > 0 {
		limit := opt.MaxResponseSize + 1
		if limit <= 0 {
			limit = opt.MaxResponseSize
		}
		limited := io.LimitReader(resp.Body, limit)
		bodyBytes, err = io.ReadAll(limited)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if int64(len(bodyBytes)) > opt.MaxResponseSize {
			_, _ = io.Copy(io.Discard, resp.Body)
			return nil, fmt.Errorf("response body exceeds max size %d bytes", opt.MaxResponseSize)
		}
	} else {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	duration := time.Since(start)

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       bodyBytes,
		RawResp:    resp,
		Duration:   duration,
	}, nil
}

// shouldRetry reports whether an error is retryable.
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "temporary failure") {
		return true
	}
	return false
}

func (r *Response) String() string {
	return string(r.Body)
}

// JSON decodes the response body as JSON into v.
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// IsSuccess reports whether the response status is 2xx.
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Get sends a GET request using the package-level client.
func Get(url string, options ...*RequestOptions) (*Response, error) {
	return Client.Get(url, options...)
}

// Post sends a POST request using the package-level client.
func Post(url string, body interface{}, options ...*RequestOptions) (*Response, error) {
	return Client.Post(url, body, options...)
}

// PostForm sends form-encoded data with merged options.
func PostForm(url string, data url.Values, options ...*RequestOptions) (*Response, error) {
	opts := &RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	}

	if len(options) > 0 && options[0] != nil {
		mergeOptions(opts, options[0])
	}

	return Post(url, data.Encode(), opts)
}

// mergeOptions merges request options.
func mergeOptions(dst, src *RequestOptions) {
	if src.Headers != nil {
		if dst.Headers == nil {
			dst.Headers = make(map[string]string)
		}
		for k, v := range src.Headers {
			dst.Headers[k] = v
		}
	}
	if src.Timeout > 0 {
		dst.Timeout = src.Timeout
	}
	if src.Proxy != nil {
		dst.Proxy = src.Proxy
	}
	if src.Context != nil {
		dst.Context = src.Context
	}
	if src.TLSConfig != nil {
		dst.TLSConfig = src.TLSConfig
	}
	if src.Jar != nil {
		dst.Jar = src.Jar
	}
	if src.DisableKeepAlives != nil {
		dst.DisableKeepAlives = src.DisableKeepAlives
	}
	if src.FollowRedirects != nil {
		dst.FollowRedirects = src.FollowRedirects
	}
	if src.MaxRedirects > 0 {
		dst.MaxRedirects = src.MaxRedirects
	}
	if src.RetryCount > 0 {
		dst.RetryCount = src.RetryCount
	}
	if src.RetryDelay > 0 {
		dst.RetryDelay = src.RetryDelay
	}
	if src.MaxResponseSize > 0 {
		dst.MaxResponseSize = src.MaxResponseSize
	}
}

// SetGlobalProxy updates the package-level client with a proxy.
func SetGlobalProxy(proxyConfig *ProxyConfig) error {
	Client.mu.RLock()
	timeout := Client.defaultTimeout
	tlsConfig := Client.defaultTLSConfig
	jar := Client.defaultJar
	disableKeepAlives := Client.defaultDisableKeepAlives
	followRedirects := Client.defaultFollowRedirects
	maxRedirects := Client.defaultMaxRedirects
	Client.mu.RUnlock()

	disableKeepAlivesPtr := disableKeepAlives
	newClient := NewHTTPClient(&RequestOptions{
		Proxy:             proxyConfig,
		Timeout:           timeout,
		TLSConfig:         tlsConfig,
		Jar:               jar,
		DisableKeepAlives: &disableKeepAlivesPtr,
		FollowRedirects:   followRedirects,
		MaxRedirects:      maxRedirects,
	})
	Client.mu.Lock()
	Client.client = newClient.client
	Client.baseTransport = newClient.baseTransport
	Client.defaultProxy = newClient.defaultProxy
	Client.defaultTimeout = newClient.defaultTimeout
	Client.defaultTLSConfig = newClient.defaultTLSConfig
	Client.defaultJar = newClient.defaultJar
	Client.defaultDisableKeepAlives = newClient.defaultDisableKeepAlives
	Client.defaultFollowRedirects = newClient.defaultFollowRedirects
	Client.defaultMaxRedirects = newClient.defaultMaxRedirects
	Client.proxyClients = newClient.proxyClients
	Client.mu.Unlock()
	return nil
}

// SetGlobalTimeout updates the package-level client timeout.
func SetGlobalTimeout(timeout time.Duration) {
	Client.mu.RLock()
	proxy := Client.defaultProxy
	tlsConfig := Client.defaultTLSConfig
	jar := Client.defaultJar
	disableKeepAlives := Client.defaultDisableKeepAlives
	followRedirects := Client.defaultFollowRedirects
	maxRedirects := Client.defaultMaxRedirects
	Client.mu.RUnlock()

	disableKeepAlivesPtr := disableKeepAlives
	newClient := NewHTTPClient(&RequestOptions{
		Proxy:             proxy,
		Timeout:           timeout,
		TLSConfig:         tlsConfig,
		Jar:               jar,
		DisableKeepAlives: &disableKeepAlivesPtr,
		FollowRedirects:   followRedirects,
		MaxRedirects:      maxRedirects,
	})
	Client.mu.Lock()
	Client.client = newClient.client
	Client.baseTransport = newClient.baseTransport
	Client.defaultProxy = newClient.defaultProxy
	Client.defaultTimeout = newClient.defaultTimeout
	Client.defaultTLSConfig = newClient.defaultTLSConfig
	Client.defaultJar = newClient.defaultJar
	Client.defaultDisableKeepAlives = newClient.defaultDisableKeepAlives
	Client.defaultFollowRedirects = newClient.defaultFollowRedirects
	Client.defaultMaxRedirects = newClient.defaultMaxRedirects
	Client.proxyClients = newClient.proxyClients
	Client.mu.Unlock()
}
