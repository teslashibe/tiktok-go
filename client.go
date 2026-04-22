package tiktok

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// waitForGap enforces the minimum request gap.
func (c *Client) waitForGap(ctx context.Context) {
	c.gapMu.Lock()
	now := time.Now()
	next := c.lastReqAt.Add(c.minGap)
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

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
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

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
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

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
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
