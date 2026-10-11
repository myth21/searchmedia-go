package searchmedia

import (
	"bytes"
	"encoding/json"
	"math"
	"net/url"
	"strings"
	"time"
)

var ImplementedServices = []string{KlipyServiceName, GoogleServiceName, GiphyServiceName}

type ServiceMeta struct {
	IsAvailable    bool    `json:"is_available"`
	ProcessingTime float64 `json:"processing_time"`
	ItemsCount     int     `json:"items_count"`
	Terms          string  `json:"terms"`
}

type namedServiceMeta struct {
	name string
	meta ServiceMeta
}

// ServicesMeta preserves service order when marshaling metadata to JSON.
type ServicesMeta []namedServiceMeta

func (m ServicesMeta) Get(serviceName string) (ServiceMeta, bool) {
	for _, entry := range m {
		if entry.name == serviceName {
			return entry.meta, true
		}
	}

	return ServiceMeta{}, false
}

func (m ServicesMeta) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteByte('{')

	for index, entry := range m {
		if index > 0 {
			buffer.WriteByte(',')
		}

		name, err := json.Marshal(entry.name)
		if err != nil {
			return nil, err
		}

		meta, err := json.Marshal(entry.meta)
		if err != nil {
			return nil, err
		}

		buffer.Write(name)
		buffer.WriteByte(':')
		buffer.Write(meta)
	}

	buffer.WriteByte('}')

	return buffer.Bytes(), nil
}

type serviceAdapter struct {
	buildURL func(config Config, query SearchQuery, limit int) string
	headers  map[string]string
	mapItems func(decodedResponse json.RawMessage) []any
	terms    string
}

var serviceAdapters = map[string]serviceAdapter{
	KlipyServiceName: {
		buildURL: func(config Config, query SearchQuery, limit int) string {
			return KlipySearchURL(config.KlipyKey, query.Text, limit, query.Offset)
		},
		headers:  klipyRequestHeaders,
		mapItems: KlipyMapItems,
		terms:    KlipyLinkToTerms,
	},
	GoogleServiceName: {
		buildURL: func(config Config, query SearchQuery, limit int) string {
			return GoogleSearchURL(config.GoogleKey, config.GoogleCX, query.Text, limit, query.Offset)
		},
		mapItems: GoogleMapItems,
		terms:    GoogleLinkToTerms,
	},
	GiphyServiceName: {
		buildURL: func(config Config, query SearchQuery, limit int) string {
			return GiphySearchURL(config.GiphyKey, query.Text, limit, query.Offset)
		},
		mapItems: GiphyMapItems,
		terms:    GiphyLinkToTerms,
	},
}

// Search returns merged results and metadata for each requested service.
func Search(client HTTPClient, config Config, query SearchQuery, availableServiceNames []string) ([]any, ServicesMeta) {
	plan := PlanForRequest(availableServiceNames, query.ServiceNames, query.Limit)

	items := []any{}
	meta := ServicesMeta{}

	for _, serviceName := range ImplementedServices {
		limit := plan.LimitFor(serviceName)
		if limit <= 0 {
			continue
		}

		adapter := serviceAdapters[serviceName]
		url := adapter.buildURL(config, query, limit)
		decodedBody, isAvailable, processingTime := callService(client, url, adapter.headers)
		serviceItems := adapter.mapItems(decodedBody)

		meta = append(meta, namedServiceMeta{
			name: serviceName,
			meta: ServiceMeta{
				IsAvailable:    isAvailable,
				ProcessingTime: processingTime,
				ItemsCount:     len(serviceItems),
				Terms:          adapter.terms,
			},
		})

		items = append(items, serviceItems...)
	}

	return items, meta
}

// Invalid responses are replaced with an empty JSON object, so no items are mapped.
func callService(client HTTPClient, url string, headers map[string]string) (json.RawMessage, bool, float64) {
	start := time.Now()
	response := client.Get(url, headers)
	processingTime := math.Round(time.Since(start).Seconds()*100) / 100

	if !response.IsOK() || !isObject(response.Body) {
		return json.RawMessage("{}"), response.IsOK(), processingTime
	}

	return response.Body, true, processingTime
}

// Unlike url.Values.Encode, encodeQuery preserves the order of the pairs.
func encodeQuery(pairs ...string) string {
	parts := make([]string, 0, len(pairs)/2)

	for index := 0; index+1 < len(pairs); index += 2 {
		parts = append(parts, url.QueryEscape(pairs[index])+"="+url.QueryEscape(pairs[index+1]))
	}

	return strings.Join(parts, "&")
}