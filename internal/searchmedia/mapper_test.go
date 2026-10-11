package searchmedia

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// readFixture reads a JSON fixture from the testdata directory.
func readFixture(t *testing.T, path ...string) json.RawMessage {
	t.Helper()

	filePath := filepath.Join(append([]string{"testdata"}, path...)...)

	body, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	return body
}

// assertSameJSON compares decoded JSON values, ignoring object key order and differences in number formatting that decode to the same value.
func assertSameJSON(t *testing.T, got any, want json.RawMessage) {
	t.Helper()

	gotBytes, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	var gotValue, wantValue any

	if err := json.Unmarshal(gotBytes, &gotValue); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Errorf("got\n%s\nwant\n%s", gotBytes, want)
	}
}

func TestMappers(t *testing.T) {
	cases := []struct {
		dir      string
		service  string
		mapItems func(json.RawMessage) []any
	}{
		{"recorded", "giphy", GiphyMapItems},
		{"recorded", "klipy", KlipyMapItems},
		{"recorded", "google", GoogleMapItems},
		// Test handling of missing and null fields in provider responses.
		{"edge-cases", "klipy", KlipyMapItems},
		{"edge-cases", "google", GoogleMapItems},
	}

	for _, c := range cases {
		t.Run(c.dir+"/"+c.service, func(t *testing.T) {
			fixture := readFixture(t, c.dir, c.service+"-search.json")
			expected := readFixture(t, c.dir, c.service+"-items.json")

			assertSameJSON(t, c.mapItems(fixture), expected)
		})
	}
}

func TestMappersSkipWrongShapes(t *testing.T) {
	testCases := map[string]string{
		"not an object":    `[]`,
		"items not a list": `{"data": "x", "results": 1, "items": true}`,
		"item not object":  `{"data": [1], "results": ["x"], "items": [null]}`,
	}

	mappers := []func(json.RawMessage) []any{
		GiphyMapItems,
		KlipyMapItems,
		GoogleMapItems,
	}

	for name, body := range testCases {
		t.Run(name, func(t *testing.T) {
			for _, mapItems := range mappers {
				items := mapItems(json.RawMessage(body))

				if len(items) != 0 {
					t.Errorf("got %d items, want none", len(items))
				}
			}
		})
	}
}

func TestIterableItemsKeepsObjectOrder(t *testing.T) {
	raw := json.RawMessage(`{"b": "first", "a": "second"}`)
	items := iterableItems(raw)
	values := decodedValues(items)

	if !reflect.DeepEqual(values, []any{"first", "second"}) {
		t.Errorf("got %v", values)
	}
}