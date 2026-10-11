package searchmedia

import (
	"encoding/json"
	"strings"
	"testing"
)

var testConfig = Config{
	SpecificationURL: "http://example.invalid/swagger/",
	DocumentationURL: "http://example.invalid/doc/",
	GiphyKey:         "giphy-key",
	KlipyKey:         "klipy-key",
	GoogleKey:        "google-key",
	GoogleCX:         "google-cx",
}

// fakeHTTPClient answers by the first host substring found in the url.
type fakeHTTPClient struct {
	responses map[string]HTTPResponse
	urls      []string
}

func (c *fakeHTTPClient) Get(url string, _ map[string]string) HTTPResponse {
	// Save requested URL so tests can check which services were called.
	c.urls = append(c.urls, url)

	// Find a prepared response for this URL.
	for host, response := range c.responses {
		if strings.Contains(url, host) {
			return response
		}
	}

	// The test did not prepare a response for this URL.
	panic("unexpected URL: " + url)
}

func fixtureResponse(t *testing.T, service string) HTTPResponse {
	return HTTPResponse{StatusCode: 200, Body: readFixture(t, "edge-cases", service+"-search.json")}
}

func fixtureItemsCount(t *testing.T, service string) int {
	var items []any
	if err := json.Unmarshal(readFixture(t, "edge-cases", service+"-items.json"), &items); err != nil {
		t.Fatal(err)
	}

	return len(items)
}

func TestSearchOneFailedServiceDoesNotAffectTheOthers(t *testing.T) {
	client := &fakeHTTPClient{responses: map[string]HTTPResponse{
		"api.klipy.com":  fixtureResponse(t, "klipy"),
		"googleapis.com": fixtureResponse(t, "google"),
		// Giphy times out: no response at all.
		"giphy.com": {},
	}}

	items, meta := Search(client, testConfig, SearchQuery{Text: "cats", Limit: 10}, ImplementedServices)

	want := map[string]ServiceMeta{
		"giphy":  {IsAvailable: false, ItemsCount: 0},
		"klipy":  {IsAvailable: true, ItemsCount: fixtureItemsCount(t, "klipy")},
		"google": {IsAvailable: true, ItemsCount: fixtureItemsCount(t, "google")},
	}

	for service, w := range want {
		got, ok := meta.Get(service)
		if !ok || got.IsAvailable != w.IsAvailable || got.ItemsCount != w.ItemsCount {
			t.Errorf("%s: got %+v", service, got)
		}
	}

	if len(items) != want["klipy"].ItemsCount+want["google"].ItemsCount {
		t.Errorf("got %d items", len(items))
	}
}

func TestSearchNon200AndNonJSONAnswers(t *testing.T) {
	client := &fakeHTTPClient{responses: map[string]HTTPResponse{
		"api.klipy.com":  {StatusCode: 500},
		"googleapis.com": {StatusCode: 200, Body: []byte("not json")},
		"giphy.com":      {StatusCode: 200, Body: []byte("{}")},
	}}

	_, meta := Search(client, testConfig, SearchQuery{Text: "cats", Limit: 10}, ImplementedServices)

	if klipy, _ := meta.Get("klipy"); klipy.IsAvailable {
		t.Error("klipy answered 500 but is available")
	}
	// Answered 200, so available, but gave nothing to map.
	if google, _ := meta.Get("google"); !google.IsAvailable || google.ItemsCount != 0 {
		t.Errorf("google: %+v", google)
	}
}

func TestSearchCallsServicesInOrderAndSkipsZeroLimits(t *testing.T) {
	client := &fakeHTTPClient{responses: map[string]HTTPResponse{
		"api.klipy.com": {StatusCode: 200, Body: []byte("{}")},
		"giphy.com":     {StatusCode: 200, Body: []byte("{}")},
	}}

	_, meta := Search(client, testConfig, SearchQuery{Text: "cats", ServiceNames: []string{"giphy", "klipy"}, Limit: 4}, ImplementedServices)

	body, _ := json.Marshal(meta)
	if !strings.HasPrefix(string(body), `{"klipy":`) || strings.Contains(string(body), "google") {
		t.Errorf("meta: %s", body)
	}
	if len(client.urls) != 2 {
		t.Errorf("called %d urls", len(client.urls))
	}
}


func TestSearchURLs(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "giphy",
			got: GiphySearchURL("k", "funny cats", 150, 5),
			want: "https://api.giphy.com/v1/gifs/search?q=funny+cats&api_key=k&limit=100&offset=5",
		},
		{
			name: "klipy",
			got: KlipySearchURL("k", "a&b", 20, 0),
			want: "https://api.klipy.com/v2/search?key=k&q=a%26b&limit=20&pos=0",
		},
		{
			name: "google",
			got: GoogleSearchURL("k", "cx", "ж", 15, 2),
			want: "https://www.googleapis.com/customsearch/v1?q=%D0%B6&searchType=image&imgSize=large&key=k&cx=cx&num=10&start=2",
		},
	}

	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got\n%s\nwant\n%s", tc.name, tc.got, tc.want)
		}
	}
}
