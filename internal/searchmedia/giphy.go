package searchmedia

import (
	"encoding/json"
	"strconv"
)

const (
	GiphyServiceName = "giphy"
	GiphySite        = "giphy.com"
	GiphyLinkToTerms = "https://support.giphy.com/hc/en-us/articles/360020027752-GIPHY-User-Terms-of-Service"

	giphySearchURL = "https://api.giphy.com/v1/gifs/search"
	giphyLimitMax  = 100
)

func GiphySearchURL(authKey, searchText string, limit, offset int) string {
	return giphySearchURL + "?" + encodeQuery(
		"q", searchText,
		"api_key", authKey,
		"limit", strconv.Itoa(min(limit, giphyLimitMax)),
		"offset", strconv.Itoa(offset),
	)
}

func GiphyMapItems(decodedResponse json.RawMessage) []any {
	items := []any{}

	for _, responseItem := range iterableItems(field(decodedResponse, "data")) {
		if !isObject(responseItem) {
			continue
		}

		url, ok := nonEmptyString(nested(responseItem, "images", "original", "url"))
		if !ok {
			continue
		}

		items = append(items, MediaItem{
			URL: url,
			// Giphy doesn't send a mime type, and search only returns gifs.
			MimeType:    "image/gif",
			Title:       stringOrEmpty(field(responseItem, "title")),
			Desc:        "",
			Tags:        []any{},
			Source:      GiphySite,
			Author:      valueOrEmpty(nested(responseItem, "user", "display_name")),
			HTMLTagName: "img",
			Type:        "image",
		})
	}

	return items
}
