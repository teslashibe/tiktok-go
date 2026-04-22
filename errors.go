package tiktok

import "errors"

var (
	ErrInvalidAuth   = errors.New("tiktok: missing required cookie (sessionid or tt_csrf_token)")
	ErrUnauthorized  = errors.New("tiktok: authentication failed — session may be expired")
	ErrForbidden     = errors.New("tiktok: access denied (private account or restricted resource)")
	ErrNotFound      = errors.New("tiktok: resource not found")
	ErrRateLimited   = errors.New("tiktok: rate limited")
	ErrEmptyResponse = errors.New("tiktok: empty response body — likely missing X-Bogus signature")
	ErrRequestFailed = errors.New("tiktok: HTTP request failed")
	ErrParseFailed   = errors.New("tiktok: failed to parse response")
)
