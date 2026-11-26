package stash

import (
	"net/url"
	"stash-vr/internal/config"
)

// PublicURL selects the client-facing origin for a resource returned by Stash.
// It deliberately leaves API key handling to the existing callers.
func PublicURL(raw string) string {
	return withOrigin(raw, config.Application().StashPublicUrl)
}

// ServerURL keeps server-side resource fetches on the GraphQL connection's
// origin, even when Stash advertises a public address in its responses.
func ServerURL(raw string) string {
	cfg := config.Application()
	if cfg.StashPublicUrl == "" {
		return raw
	}
	return withOrigin(raw, cfg.StashGraphQLUrl)
}

func withOrigin(raw, origin string) string {
	if origin == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Opaque != "" || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") {
		return raw
	}
	base, err := url.Parse(origin)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return raw
	}
	u.Scheme, u.Host, u.User = base.Scheme, base.Host, nil
	return u.String()
}
