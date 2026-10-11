package searchmedia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func serve(t *testing.T, handler http.Handler, target string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s: body is not JSON: %s", target, recorder.Body.String())
	}

	return recorder, body
}

func newTestHandler(client HTTPClient) http.Handler {
	return NewHandler(testConfig, client, NewRequestThrottle(OneRequestIn))
}

func TestServiceNamesEndpoint(t *testing.T) {
	recorder, body := serve(t, newTestHandler(&fakeHTTPClient{}), ServiceNamesPath)

	if recorder.Code != 200 || recorder.Header().Get("Content-Type") != jsonMediaType {
		t.Fatalf("got %d %s", recorder.Code, recorder.Header().Get("Content-Type"))
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(recorder.Header().Get(requestIDHeader)) {
		t.Errorf("request id header: %q", recorder.Header().Get(requestIDHeader))
	}
	if body["request_id"] != recorder.Header().Get(requestIDHeader) {
		t.Error("request_id in the body differs from the header")
	}

	names, _ := json.Marshal(body["data"])
	if string(names) != `["klipy","google","giphy"]` {
		t.Errorf("data: %s", names)
	}
	for _, key := range []string{"meta", "pagination"} {
		if value, _ := json.Marshal(body[key]); string(value) != "{}" {
			t.Errorf("%s: %s", key, value)
		}
	}
	if !regexp.MustCompile(`^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{6}\+00:00$`).MatchString(body["timestamp"].(string)) {
		t.Errorf("timestamp: %v", body["timestamp"])
	}
}

func TestSearchEndpoint(t *testing.T) {
	client := &fakeHTTPClient{responses: map[string]HTTPResponse{
		"api.klipy.com":  fixtureResponse(t, "klipy"),
		"googleapis.com": fixtureResponse(t, "google"),
		"giphy.com":      {},
	}}

	recorder, body := serve(t, newTestHandler(client), SearchPath+"?text=cats&limit=9&offset=3")

	if recorder.Code != 200 {
		t.Fatalf("got %d: %s", recorder.Code, recorder.Body.String())
	}

	pagination, _ := json.Marshal(body["pagination"])
	if string(pagination) != `{"items_count":4,"limit":9,"offset":3}` {
		t.Errorf("pagination: %s", pagination)
	}
	if data, _ := body["data"].([]any); len(data) != 4 {
		t.Errorf("data: %v", body["data"])
	}
	if _, ok := body["error"]; ok {
		t.Error("a successful answer carries error")
	}
}

func TestSearchEndpointWithNoItemsStillHasADataList(t *testing.T) {
	client := &fakeHTTPClient{responses: map[string]HTTPResponse{"giphy.com": {}}}

	_, body := serve(t, newTestHandler(client), SearchPath+"?text=cats&service=giphy")

	if data, ok := body["data"].([]any); !ok || len(data) != 0 {
		t.Errorf("data: %#v", body["data"])
	}
}

func TestSearchEndpointBadRequest(t *testing.T) {
	recorder, body := serve(t, newTestHandler(&fakeHTTPClient{}), SearchPath)

	if recorder.Code != 400 || body["status_text"] != "Bad Request" {
		t.Fatalf("got %d %v", recorder.Code, body["status_text"])
	}

	errorBody, _ := json.Marshal(body["error"])
	want := `{"details":[],"help":"http://example.invalid/swagger/","id":"","message":"The ` + "`text`" + ` parameter is missing."}`
	if string(errorBody) != want {
		t.Errorf("error: %s", errorBody)
	}
}

func TestSearchEndpointThrottlesByIP(t *testing.T) {
	handler := newTestHandler(&fakeHTTPClient{})

	serve(t, handler, SearchPath)
	recorder, body := serve(t, handler, SearchPath)

	if recorder.Code != 429 {
		t.Fatalf("got %d", recorder.Code)
	}


	rawError := body["error"]
	errorBody, _ := rawError.(map[string]any)

	if errorBody["message"] != "Too Many Requests" || errorBody["help"] != "http://example.invalid/doc/" {
		t.Errorf("error: %v", errorBody)
	}

	// The service names endpoint is not throttled.
	if recorder, _ := serve(t, handler, ServiceNamesPath); recorder.Code != 200 {
		t.Errorf("service names after throttling: %d", recorder.Code)
	}
}

func TestResponseDoesNotEscapeHTMLCharacters(t *testing.T) {
	client := &fakeHTTPClient{responses: map[string]HTTPResponse{"giphy.com": {
		StatusCode: 200,
		Body:       []byte(`{"data":[{"title":"a<b>&c","images":{"original":{"url":"https://x/y.gif?a=1&b=2"}}}]}`),
	}}}

	recorder, _ := serve(t, newTestHandler(client), SearchPath+"?text=cats&service=giphy")

	if !regexp.MustCompile(`"title":"a<b>&c"`).MatchString(recorder.Body.String()) {
		t.Errorf("body: %s", recorder.Body.String())
	}
	if last := recorder.Body.Bytes()[recorder.Body.Len()-1]; last != '}' {
		t.Errorf("body ends with %q", last)
	}
}
