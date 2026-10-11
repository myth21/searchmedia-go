package searchmedia

// MediaItem is one element of data[] for Giphy and Klipy. Field order follows the other implementations.
type MediaItem struct {
	URL         string `json:"url"`
	MimeType    string `json:"mime_type"`
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Tags        []any  `json:"tags"`
	Source      string `json:"source"`
	Author      any    `json:"author"`
	HTMLTagName string `json:"html_tag_name"`
	Type        string `json:"type"`
}

// GoogleItem also carries fallback_url, always present even when null, and puts author before source.
type GoogleItem struct {
	URL         string `json:"url"`
	FallbackURL any    `json:"fallback_url"`
	MimeType    string `json:"mime_type"`
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Tags        []any  `json:"tags"`
	Author      any    `json:"author"`
	Source      string `json:"source"`
	HTMLTagName string `json:"html_tag_name"`
	Type        string `json:"type"`
}
