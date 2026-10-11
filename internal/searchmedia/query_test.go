package searchmedia

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

const maxLimit = 30

func readTestQuery(t *testing.T, raw string) (SearchQuery, error) {
	t.Helper()

	params, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}

	return ReadQuery(params, AvailableServiceNames, maxLimit)
}

func TestReadQueryRejects(t *testing.T) {
	cases := map[string]string{
		"missing text":          "",
		"repeated text":         "text=cats&text=dogs",
		"blank text":            "text=%20%20%20",
		"text over 32 chars":    "text=" + strings.Repeat("a", 33),
		"bracketed text":        "text[]=cats",
		"unknown service":       "text=cats&service=imgur",
		"zero limit":            "text=cats&limit=0",
		"non-integer limit":     "text=cats&limit=abc",
		"negative limit":        "text=cats&limit=-5",
		"plus-signed limit":     "text=cats&limit=%2B5",
		"non-ascii digit limit": "text=cats&limit=%C2%B2",
		"empty limit":           "text=cats&limit=",
		"repeated limit":        "text=cats&limit=5&limit=6",
		"non-integer offset":    "text=cats&offset=abc",
		"huge offset":           "text=cats&offset=99999999999999999999",
	}

	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := readTestQuery(t, raw)

			var invalidQuery *InvalidSearchQuery
			// If err does NOT contain an error of type *InvalidSearchQuery, then execute the block.
			if !errors.As(err, &invalidQuery) {
				t.Errorf("got %v, want InvalidSearchQuery", err)
			}
		})
	}
}

func TestReadQueryMessages(t *testing.T) {
	cases := map[string]string{
		"":                                "The `text` parameter is missing.",
		"text=a&text=b":                   "The `text` parameter must be given once.",
		"text=%20%20":                     "The `text` parameter is empty.",
		"text=" + strings.Repeat("a", 33): "The `text` must be between 1 and 32 characters long.",
		"text=cats&service=nope":          "The value of the `service` parameter must be one of the allowed values: klipy, google, giphy.",
		"text=cats&limit=0":               "The value of the `limit` parameter must be greater than 0.",
		"text=cats&limit=x":               "The value of the `limit` parameter must be an integer.",
		"text=cats&offset=1&offset=2":     "The `offset` parameter must be given once.",
	}

	for raw, want := range cases {
		_, err := readTestQuery(t, raw)
		if err == nil || err.Error() != want {
			t.Errorf("%q: got %v, want %q", raw, err, want)
		}
	}
}

func TestReadQueryDefaultsAndValues(t *testing.T) {
	cases := map[string]SearchQuery{
		"text=%20%20cats%20%20":                {Text: "cats", Limit: maxLimit},
		"text=cats&limit=1000":                 {Text: "cats", Limit: maxLimit},
		"text=cats&limit=99999999999999999999": {Text: "cats", Limit: maxLimit},
		"text=cats&limit=5":                    {Text: "cats", Limit: 5},
		"text=cats&offset=10":                  {Text: "cats", Limit: maxLimit, Offset: 10},
		"text=cats&service=klipy":              {Text: "cats", ServiceNames: []string{"klipy"}, Limit: maxLimit},
		"text=cats&service=klipy&service=google&limit=7&offset=3": {
			Text: "cats", ServiceNames: []string{"klipy", "google"}, Limit: 7, Offset: 3,
		},
		// 32 characters, counted as characters rather than bytes.
		"text=" + strings.Repeat("%D0%B6", 32): {Text: strings.Repeat("ж", 32), Limit: maxLimit},
	}

	for raw, want := range cases {
		got, err := readTestQuery(t, raw)
		if err != nil {
			t.Errorf("%q: unexpected error %v", raw, err)
			continue
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %+v, want %+v", raw, got, want)
		}
	}
}
