package deovr

import (
	"encoding/json"
	"flag"
	"github.com/spf13/pflag"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/stash/gql"
	"strings"
	"testing"
)

func TestPublicMediaURLs(t *testing.T) {
	t.Setenv("STASH_GRAPHQL_URL", "http://internal:9999/graphql")
	t.Setenv("STASH_PUBLIC_URL", "https://domain.com")
	t.Setenv("STASH_API_KEY", "test-key")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	config.Init()
	var scene gql.SceneParts
	if err := json.Unmarshal([]byte(`{
 "id":"1","title":"Test","files":[{"height":1080,"basename":"test.mp4"}],
 "paths":{"screenshot":"http://internal:9999/scene/1/screenshot?x=1","preview":"http://internal:9999/scene/1/preview","stream":"http://internal:9999/scene/1/stream?apikey=existing","funscript":"http://internal:9999/scene/1/funscript","caption":"http://internal:9999/scene/1/caption"},
 "sceneStreams":[{"url":"http://internal:9999/scene/1/stream.mp4?apikey=existing","mime_type":"video/mp4","label":"MP4 (1080p)"}],
 "interactive":true,"captions":[{"language_code":"en","caption_type":"srt"}]
}`), &scene); err != nil {
		t.Fatal(err)
	}
	vd := &library.VideoData{SceneParts: &scene}
	dto, err := buildVideoData(vd, "https://wrapper.example")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	// Decode JSON string escapes before inspecting the URLs.
	text := strings.ReplaceAll(string(data), `\u0026`, "&")
	if strings.Contains(text, "http://internal") {
		t.Fatalf("internal URL exposed: %s", text)
	}
	for _, suffix := range []string{"screenshot?x=1&apikey=test-key", "preview?apikey=test-key", "stream?apikey=existing", "stream.mp4?apikey=existing"} {
		if !strings.Contains(text, "https://domain.com/scene/1/"+suffix) {
			t.Errorf("missing public resource %s in %s", suffix, text)
		}
	}
	if *scene.Paths.Screenshot != "http://internal:9999/scene/1/screenshot?x=1" {
		t.Error("cached server URL was mutated")
	}
	index, err := buildIndex([]library.Section{{Ids: []string{"1"}}}, map[string]*library.VideoData{"1": vd}, "https://wrapper.example")
	if err != nil {
		t.Fatal(err)
	}
	if got := *index.Scenes[0].List[0].ThumbnailUrl; got != "https://domain.com/scene/1/screenshot?x=1&apikey=test-key" {
		t.Fatalf("index thumbnail: %s", got)
	}

}
