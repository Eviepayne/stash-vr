package config

import "testing"

func TestValidateSplitURLs(t *testing.T) {
	for _, tc := range []struct {
		public, backend string
		valid           bool
	}{
		{"", "", true},
		{"https://domain.com", "http://localhost:9999/graphql", true},
		{"https://domain.com:443/", "http://[::1]:9999/stash/graphql", true},
		{"domain.com", "http://localhost/graphql", false},
		{"https://domain.com/graphql", "http://localhost/graphql", false},
		{"https://domain.com?key=x", "http://localhost/graphql", false},
		{"https://user:secret@domain.com", "http://localhost/graphql", false},
		{"https://domain.com", "ftp://localhost/graphql", false},
		{"https://domain.com", "http://%zz", false},
	} {
		cfg := ApplicationConfig{StashPublicUrl: tc.public, StashGraphQLUrl: tc.backend}
		if err := cfg.Validate(); (err == nil) != tc.valid {
			t.Errorf("public=%q backend=%q: err=%v", tc.public, tc.backend, err)
		}
	}
}
