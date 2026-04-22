package tiktok

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// commonParams returns the standard query parameters sent on every API call.
func (c *Client) commonParams() url.Values {
	v := url.Values{}
	v.Set("aid", "1988")
	v.Set("app_name", "tiktok_web")
	v.Set("device_platform", "web_pc")
	v.Set("browser_language", "en-US")
	v.Set("browser_platform", "MacIntel")
	v.Set("browser_name", "Mozilla")
	v.Set("browser_version", "5.0")
	v.Set("region", "US")
	v.Set("language", "en")
	v.Set("os", "mac")
	v.Set("screen_height", "1080")
	v.Set("screen_width", "1920")
	v.Set("tz_name", "America/New_York")
	v.Set("cookie_enabled", "1")
	v.Set("focus_state", "true")
	v.Set("is_fullscreen", "false")
	v.Set("is_page_visible", "true")
	v.Set("history_len", "3")
	if ms := c.msToken(); ms != "" {
		v.Set("msToken", ms)
	}
	return v
}

// setAPIHeaders sets the required headers for an API JSON request.
func (c *Client) setAPIHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cookie", c.buildCookieHeader())
	req.Header.Set("Referer", referer)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

// setPageHeaders sets headers for an HTML page scrape.
func (c *Client) setPageHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cookie", c.buildCookieHeader())
	req.Header.Set("Referer", referer)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

// setWriteHeaders sets extra headers required for write (POST) operations.
func (c *Client) setWriteHeaders(req *http.Request, referer string) {
	c.setAPIHeaders(req, referer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Secsdk-Csrf-Version", "1.2.8")
	req.Header.Set("X-Tt-Csrf-Token", c.cookies.CSRFToken)
}

// waitForGap enforces the minimum request gap using adaptive rate-limit state.
func (c *Client) waitForGap(ctx context.Context) {
	gap := c.adaptiveGap()

	c.gapMu.Lock()
	now := time.Now()
	next := c.lastReqAt.Add(gap)
	if now.After(next) {
		next = now
	}
	c.lastReqAt = next
	c.gapMu.Unlock()

	if wait := time.Until(next); wait > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
	}
	// Clear RetryAfter once we've waited past it.
	c.rlMu.Lock()
	c.rlState.RetryAfter = 0
	c.rlMu.Unlock()
}

// adaptiveGap returns the delay before the next request based on observed
// rate-limit state. Spreads requests across the window when quota is low;
// waits for reset when quota is exhausted.
func (c *Client) adaptiveGap() time.Duration {
	c.rlMu.Lock()
	rs := c.rlState
	c.rlMu.Unlock()

	// Quota exhausted — wait for the window to reset.
	if rs.Remaining == 0 && !rs.Reset.IsZero() {
		if d := time.Until(rs.Reset); d > 0 {
			return d + 50*time.Millisecond
		}
	}
	// Spread remaining quota evenly across the reset window (90% safety margin).
	if rs.Remaining > 0 && !rs.Reset.IsZero() {
		if d := time.Until(rs.Reset); d > 0 {
			spread := d / time.Duration(float64(rs.Remaining)*0.9)
			if spread > c.minGap {
				return spread
			}
		}
	}
	return c.minGap
}

// updateRateLimit reads standard rate-limit headers from a response and updates
// the client's tracked state. Call on every HTTP response.
func (c *Client) updateRateLimit(h http.Header) {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()
	if v := rlHeader(h, "Limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.rlState.Limit = n
		}
	}
	if v := rlHeader(h, "Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.rlState.Remaining = n
		}
	}
	if v := rlHeader(h, "Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			if ts > 1_000_000_000 {
				c.rlState.Reset = time.Unix(ts, 0) // Unix epoch (Twitter/X style)
			} else {
				c.rlState.Reset = time.Now().Add(time.Duration(ts) * time.Second) // relative (Reddit style)
			}
		}
	}
}

// rlHeader returns the trimmed value of a rate-limit header, checking the four
// most common prefix variants.
func rlHeader(h http.Header, suffix string) string {
	for _, p := range []string{"X-RateLimit-", "X-Rate-Limit-", "X-Ratelimit-", "RateLimit-"} {
		if v := strings.TrimSpace(h.Get(p + suffix)); v != "" {
			return v
		}
	}
	return ""
}

// parseRetryAfter parses rate-limit headers. Handles three formats:
// - Seconds integer (Retry-After: 60)
// - Unix epoch timestamp (X-Rate-Limit-Reset: 1716000000)
// - HTTP-date (Retry-After: Mon, 01 Jan 2024 00:00:00 GMT)
func parseRetryAfter(val string, fallback time.Duration) time.Duration {
	if val == "" {
		return fallback
	}
	trimmed := strings.TrimSpace(val)
	if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		if n > 1_000_000_000 {
			if d := time.Until(time.Unix(n, 0)); d > 0 {
				return d
			}
			return fallback
		}
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(trimmed); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return fallback
}

// apiGET performs an authenticated GET request to a TikTok API endpoint.
// It appends standard params and the X-Bogus signature automatically.
func (c *Client) apiGET(ctx context.Context, path string, extra url.Values, referer string) ([]byte, error) {
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	params := c.commonParams()
	for k, vs := range extra {
		for _, v := range vs {
			params.Set(k, v)
		}
	}

	queryStr := params.Encode()
	xb := calcXBogus(queryStr, c.userAgent)
	params.Set("X-Bogus", xb)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	c.setAPIHeaders(req, referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	c.updateMsToken(resp)
	c.updateRateLimit(resp.Header)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		c.rlMu.Lock()
		c.rlState.Remaining = 0
		c.rlState.RetryAfter = wait
		if c.rlState.Reset.IsZero() || time.Until(c.rlState.Reset) < wait {
			c.rlState.Reset = time.Now().Add(wait)
		}
		c.rlMu.Unlock()
		c.gapMu.Lock()
		if earliest := time.Now().Add(wait); c.lastReqAt.Before(earliest) {
			c.lastReqAt = earliest
		}
		c.gapMu.Unlock()
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrRequestFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}
	if len(body) == 0 {
		return nil, ErrEmptyResponse
	}
	return body, nil
}

// apiGETNoSign performs a GET without adding X-Bogus (for Tier 1 endpoints).
func (c *Client) apiGETNoSign(ctx context.Context, path string, extra url.Values, referer string) ([]byte, error) {
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	params := c.commonParams()
	for k, vs := range extra {
		for _, v := range vs {
			params.Set(k, v)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	c.setAPIHeaders(req, referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	c.updateMsToken(resp)
	c.updateRateLimit(resp.Header)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		c.rlMu.Lock()
		c.rlState.Remaining = 0
		c.rlState.RetryAfter = wait
		if c.rlState.Reset.IsZero() || time.Until(c.rlState.Reset) < wait {
			c.rlState.Reset = time.Now().Add(wait)
		}
		c.rlMu.Unlock()
		c.gapMu.Lock()
		if earliest := time.Now().Add(wait); c.lastReqAt.Before(earliest) {
			c.lastReqAt = earliest
		}
		c.gapMu.Unlock()
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrRequestFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}
	if len(body) == 0 {
		return nil, ErrEmptyResponse
	}
	return body, nil
}

// apiPOST performs an authenticated form-encoded POST to a TikTok API endpoint.
// X-Bogus is appended to the URL query string.
func (c *Client) apiPOST(ctx context.Context, path string, form url.Values, referer string) ([]byte, error) {
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	params := c.commonParams()
	queryStr := params.Encode()
	xb := calcXBogus(queryStr, c.userAgent)
	params.Set("X-Bogus", xb)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+path+"?"+params.Encode(),
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	c.setWriteHeaders(req, referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	c.updateMsToken(resp)
	c.updateRateLimit(resp.Header)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		c.rlMu.Lock()
		c.rlState.Remaining = 0
		c.rlState.RetryAfter = wait
		if c.rlState.Reset.IsZero() || time.Until(c.rlState.Reset) < wait {
			c.rlState.Reset = time.Now().Add(wait)
		}
		c.rlMu.Unlock()
		c.gapMu.Lock()
		if earliest := time.Now().Add(wait); c.lastReqAt.Before(earliest) {
			c.lastReqAt = earliest
		}
		c.gapMu.Unlock()
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrRequestFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}
	return body, nil
}

// apiPOSTNoSign performs a POST without adding X-Bogus to the URL (for Tier 1 write endpoints like comments).
func (c *Client) apiPOSTNoSign(ctx context.Context, path string, form url.Values, referer string) ([]byte, error) {
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	params := c.commonParams()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+path+"?"+params.Encode(),
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	c.setWriteHeaders(req, referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	c.updateMsToken(resp)
	c.updateRateLimit(resp.Header)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		c.rlMu.Lock()
		c.rlState.Remaining = 0
		c.rlState.RetryAfter = wait
		if c.rlState.Reset.IsZero() || time.Until(c.rlState.Reset) < wait {
			c.rlState.Reset = time.Now().Add(wait)
		}
		c.rlMu.Unlock()
		c.gapMu.Lock()
		if earliest := time.Now().Add(wait); c.lastReqAt.Before(earliest) {
			c.lastReqAt = earliest
		}
		c.gapMu.Unlock()
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrRequestFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}
	return body, nil
}

// pageGET fetches an HTML page for SSR data extraction.
func (c *Client) pageGET(ctx context.Context, path string, referer string) ([]byte, error) {
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	c.setPageHeaders(req, referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	c.updateMsToken(resp)
	c.updateRateLimit(resp.Header)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		c.rlMu.Lock()
		c.rlState.Remaining = 0
		c.rlState.RetryAfter = wait
		if c.rlState.Reset.IsZero() || time.Until(c.rlState.Reset) < wait {
			c.rlState.Reset = time.Now().Add(wait)
		}
		c.rlMu.Unlock()
		c.gapMu.Lock()
		if earliest := time.Now().Add(wait); c.lastReqAt.Before(earliest) {
			c.lastReqAt = earliest
		}
		c.gapMu.Unlock()
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrRequestFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}
	return body, nil
}
