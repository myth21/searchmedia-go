package searchmedia

import (
	"encoding/json"
	"strconv"
)

const (
	KlipyServiceName = "klipy"
	KlipySite        = "api.klipy.com"
	KlipyLinkToTerms = "https://klipy.com/support/terms-services"

	klipySearchURL = "https://api.klipy.com/v2/search"
)

// Only Klipy requires a request header.
var klipyRequestHeaders = map[string]string{"Content-Type": "application/json"}

func KlipySearchURL(authKey, searchText string, limit, offset int) string {
	return klipySearchURL + "?" + encodeQuery(
		"key", authKey,
		"q", searchText,
		"limit", strconv.Itoa(limit),
		"pos", strconv.Itoa(offset),
	)
}

func KlipyMapItems(decodedResponse json.RawMessage) []any {
	items := []any{}

	for _, responseItem := range iterableItems(field(decodedResponse, "results")) {
		if !isObject(responseItem) {
			continue
		}

		url, ok := nonEmptyString(nested(responseItem, "media_formats", "gif", "url"))
		if !ok {
			continue
		}

		var author any = ""
		if user := field(responseItem, "user"); isObject(user) {
			author = valueOrEmpty(field(user, "profile_url"))
		}

		items = append(items, MediaItem{
			URL:         url,
			MimeType:    "image/gif",
			Title:       stringOrEmpty(field(responseItem, "title")),
			Desc:        stringOrEmpty(field(responseItem, "content_description")),
			Tags:        decodedValues(iterableItems(field(responseItem, "tags"))),
			Source:      KlipySite,
			Author:      author,
			HTMLTagName: "img",
			Type:        "image",
		})
	}

	return items
}
