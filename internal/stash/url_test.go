package stash

import (
	"context"
	"flag"
	"github.com/spf13/pflag"
	"net/http"
	"net/http/httptest"
	"stash-vr/internal/config"
	"strings"
	"testing"
)

func TestWithOrigin(t *testing.T) {
	for _, tc := range []struct{ name, raw, origin, want string }{
		{"disabled", "http://internal:9999/scene/1?apikey=x", "", "http://internal:9999/scene/1?apikey=x"},
		{"public TLS", "http://internal:9999/scene/a%2Fb?apikey=x%2By&n=1#part", "https://domain.com/", "https://domain.com/scene/a%2Fb?apikey=x%2By&n=1#part"},
		{"server HTTP", "https://domain.com/stash/scene/1?apikey=x", "http://127.0.0.1:9999/stash/graphql", "http://127.0.0.1:9999/stash/scene/1?apikey=x"},
		{"relative", "/scene/1", "https://domain.com", "https://domain.com/scene/1"},
		{"IPv6", "http://old/scene/1", "http://[::1]:9999/graphql", "http://[::1]:9999/scene/1"},
		{"invalid resource", "http://%zz", "https://domain.com", "http://%zz"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := withOrigin(tc.raw, tc.origin); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestURLPerspectives(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			t.Errorf("unexpected backend path %q", r.URL.Path)
		}
		if r.Header.Get("ApiKey") != "test-key" {
			t.Error("missing backend API key")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"version":{"version":"test-version"}}}`))
	}))
	defer server.Close()
	raw := "https://advertised.example/scene/1?apikey=existing"
	if PublicURL(raw) != raw || ServerURL(raw) != raw {
		t.Fatal("unset public origin changed existing behavior")
	}
	t.Setenv("STASH_GRAPHQL_URL", server.URL+"/graphql")
	t.Setenv("STASH_PUBLIC_URL", "https://domain.com")
	t.Setenv("STASH_API_KEY", "test-key")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	config.Init()
	if err := config.Application().Validate(); err != nil {
		t.Fatal(err)
	}
	version, err := GetVersion(context.Background(), NewClient(config.Application().StashGraphQLUrl, config.Application().StashApiKey))
	if err != nil || version != "test-version" {
		t.Fatalf("backend query failed: version=%q err=%v", version, err)
	}
	if got := PublicURL(raw); got != "https://domain.com/scene/1?apikey=existing" {
		t.Fatalf("public URL: %s", got)
	}
	if got := ServerURL(raw); got != server.URL+"/scene/1?apikey=existing" {
		t.Fatalf("server URL: %s", got)
	}
	if got := ApiKeyed(PublicURL(raw)); strings.Count(got, "apikey=") != 1 {
		t.Fatalf("duplicated key: %s", got)
	}
}
