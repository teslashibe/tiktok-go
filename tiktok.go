package tiktok

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	baseURL          = "https://www.tiktok.com"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	defaultMinGap    = 500 * time.Millisecond
	maxResponseBody  = 10 << 20 // 10 MB
)

// Cookies holds TikTok session cookies from a browser export.
// SessionID and CSRFToken are required; all others are strongly recommended.
type Cookies struct {
	SessionID string // sessionid — primary auth
	SIDtt     string // sid_tt
	CSRFToken string // tt_csrf_token — also sent as X-Tt-Csrf-Token header
	MsToken   string // msToken — rotates on every response
	TTWid     string // ttwid
	OdinTT    string // odin_tt
	SIDUcpV1  string // sid_ucp_v1
	UIDtt     string // uid_tt
	// Optional extras for maximum compatibility
	Ttp          string // _ttp
	SVWebID      string // s_v_web_id
	DTicket      string // d_ticket
	SidGuard     string // sid_guard
	UIDttSS      string // uid_tt_ss
	SessionIDSS  string // sessionid_ss
	TTChainToken string // tt_chain_token
}

// Client is a TikTok web API client. Safe for concurrent use.
type Client struct {
	cookies    Cookies
	httpClient *http.Client
	userAgent  string
	minGap     time.Duration
	gapMu      sync.Mutex
	lastReqAt  time.Time
	msMu       sync.Mutex // protects msToken rotation
	rlMu       sync.Mutex
	rlState    RateLimitState
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the default http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithUserAgent overrides the default Chrome User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithMinRequestGap sets the minimum time between consecutive requests.
// Default: 500ms.
func WithMinRequestGap(d time.Duration) Option {
	return func(c *Client) { c.minGap = d }
}

// New creates a Client. SessionID and CSRFToken are required.
func New(cookies Cookies, opts ...Option) (*Client, error) {
	if cookies.SessionID == "" || cookies.CSRFToken == "" {
		return nil, fmt.Errorf("%w: SessionID and CSRFToken must both be non-empty", ErrInvalidAuth)
	}
	c := &Client{
		cookies:    cookies,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  defaultUserAgent,
		minGap:     defaultMinGap,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// RateLimit returns a snapshot of the most recently observed rate-limit state.
// Use RateLimitState.IsLimited() to check if the client is currently throttled.
func (c *Client) RateLimit() RateLimitState {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()
	return c.rlState
}

// buildCookieHeader constructs the Cookie header value from the Cookies struct.
func (c *Client) buildCookieHeader() string {
	c.msMu.Lock()
	ms := c.cookies.MsToken
	c.msMu.Unlock()

	pairs := []struct{ k, v string }{
		{"sessionid", c.cookies.SessionID},
		{"sid_tt", c.cookies.SIDtt},
		{"tt_csrf_token", c.cookies.CSRFToken},
		{"msToken", ms},
		{"ttwid", c.cookies.TTWid},
		{"odin_tt", c.cookies.OdinTT},
		{"sid_ucp_v1", c.cookies.SIDUcpV1},
		{"uid_tt", c.cookies.UIDtt},
		{"_ttp", c.cookies.Ttp},
		{"s_v_web_id", c.cookies.SVWebID},
		{"d_ticket", c.cookies.DTicket},
		{"sid_guard", c.cookies.SidGuard},
		{"uid_tt_ss", c.cookies.UIDttSS},
		{"sessionid_ss", c.cookies.SessionIDSS},
		{"tt_chain_token", c.cookies.TTChainToken},
	}
	var b strings.Builder
	for _, p := range pairs {
		if p.v == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		b.WriteString(p.k)
		b.WriteByte('=')
		b.WriteString(p.v)
	}
	return b.String()
}

// updateMsToken reads Set-Cookie from a response and rotates msToken.
func (c *Client) updateMsToken(resp *http.Response) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "msToken" && cookie.Value != "" {
			c.msMu.Lock()
			c.cookies.MsToken = cookie.Value
			c.msMu.Unlock()
			return
		}
	}
}

// msToken returns the current msToken value.
func (c *Client) msToken() string {
	c.msMu.Lock()
	defer c.msMu.Unlock()
	return c.cookies.MsToken
}
