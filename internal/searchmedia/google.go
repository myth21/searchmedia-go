package searchmedia

import (
	"encoding/json"
	"strconv"
)

const (
	GoogleServiceName = "google"
	GoogleSite        = "googleapis.com"
	GoogleLinkToTerms = "https://developers.google.com/terms"

	googleSearchURL = "https://www.googleapis.com/customsearch/v1"
	// Custom Search returns at most 10 results, regardless of num.
	googleLimitMax = 10
)

func GoogleSearchURL(authKey, clientKey, searchText string, limit, offset int) string {
	return googleSearchURL + "?" + encodeQuery(
		"q", searchText,
		"searchType", "image",
		"imgSize", "large",
		"key", authKey,
		"cx", clientKey,
		"num", strconv.Itoa(min(limit, googleLimitMax)),
		"start", strconv.Itoa(offset),
	)
}

func GoogleMapItems(decodedResponse json.RawMessage) []any {
	items := []any{}

	for _, responseItem := range iterableItems(field(decodedResponse, "items")) {
		if !isObject(responseItem) {
			continue
		}

		// The original image first, Google thumbnail as a fallback.
		url, ok := nonEmptyString(field(responseItem, "link"))
		if !ok {
			continue
		}

		items = append(items, GoogleItem{
			URL:         url,
			FallbackURL: valueOrNull(nested(responseItem, "image", "thumbnailLink")),
			MimeType:    stringOrEmpty(field(responseItem, "mime")),
			Title:       stringOrEmpty(field(responseItem, "title")),
			Desc:        stringOrEmpty(field(responseItem, "snippet")),
			Tags:        []any{},
			Author:      valueOrEmpty(field(responseItem, "displayLink")),
			Source:      GoogleSite,
			HTMLTagName: "img",
			Type:        "image",
		})
	}

	return items
}
