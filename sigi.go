package tiktok

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var sigiRe = regexp.MustCompile(`(?s)<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__"[^>]*>(.*?)</script>`)

// extractSIGI parses the __UNIVERSAL_DATA_FOR_REHYDRATION__ JSON blob
// embedded in TikTok SSR pages and returns the __DEFAULT_SCOPE__ map.
func extractSIGI(html []byte) (map[string]json.RawMessage, error) {
	m := sigiRe.FindSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("%w: __UNIVERSAL_DATA_FOR_REHYDRATION__ not found in page", ErrParseFailed)
	}

	var root struct {
		DefaultScope map[string]json.RawMessage `json:"__DEFAULT_SCOPE__"`
	}
	if err := json.Unmarshal(m[1], &root); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}
	if root.DefaultScope == nil {
		return nil, fmt.Errorf("%w: __DEFAULT_SCOPE__ is null", ErrParseFailed)
	}
	return root.DefaultScope, nil
}
