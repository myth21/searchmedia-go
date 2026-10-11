package searchmedia

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	MinSearchTextLength = 1
	MaxSearchTextLength = 32
)

// InvalidSearchQuery's message goes into error.message of the response.
type InvalidSearchQuery struct {
	message string
}

func (e *InvalidSearchQuery) Error() string {
	return e.message
}

func invalid(format string, args ...any) error {
	return &InvalidSearchQuery{message: fmt.Sprintf(format, args...)}
}

type SearchQuery struct {
	Text string
	// Nil means all available services.
	ServiceNames []string
	Limit        int
	Offset       int
}

func ReadQuery(params url.Values, availableServiceNames []string, maxLimit int) (SearchQuery, error) {
	text, err := readText(params)
	if err != nil {
		return SearchQuery{}, err
	}

	serviceNames, err := readServiceNames(params, availableServiceNames)
	if err != nil {
		return SearchQuery{}, err
	}

	limit, err := readLimit(params, maxLimit)
	if err != nil {
		return SearchQuery{}, err
	}

	offset, err := readOffset(params)
	if err != nil {
		return SearchQuery{}, err
	}

	return SearchQuery{Text: text, ServiceNames: serviceNames, Limit: limit, Offset: offset}, nil
}

func readText(params url.Values) (string, error) {
	given := params["text"]

	if len(given) == 0 {
		return "", invalid("The `text` parameter is missing.")
	}

	if len(given) > 1 {
		return "", invalid("The `text` parameter must be given once.")
	}

	text := strings.TrimSpace(given[0])

	if text == "" {
		return "", invalid("The `text` parameter is empty.")
	}

	if length := utf8.RuneCountInString(text); length < MinSearchTextLength || length > MaxSearchTextLength {
		return "", invalid(
			"The `text` must be between %d and %d characters long.",
			MinSearchTextLength, MaxSearchTextLength,
		)
	}

	return text, nil
}

func readServiceNames(params url.Values, availableServiceNames []string) ([]string, error) {
	requested := params["service"]

	if len(requested) == 0 {
		return nil, nil
	}

	for _, name := range requested {
		if !slices.Contains(availableServiceNames, name) {
			return nil, invalid(
				"The value of the `service` parameter must be one of the allowed values: %s.",
				strings.Join(availableServiceNames, ", "),
			)
		}
	}

	return requested, nil
}

func readLimit(params url.Values, maxLimit int) (int, error) {
	raw, err := readSingleValue(params, "limit")
	if err != nil || raw == nil {
		return maxLimit, err
	}

	value, err := readInteger(*raw, "limit")
	if err != nil {
		// Only digits get here, so the only failure is a number too big for int, which the cap below would cut anyway.
		if _, tooBig := err.(*strconv.NumError); tooBig {
			return maxLimit, nil
		}

		return 0, err
	}

	if value < 1 {
		return 0, invalid("The value of the `limit` parameter must be greater than 0.")
	}

	return min(value, maxLimit), nil
}

func readOffset(params url.Values) (int, error) {
	raw, err := readSingleValue(params, "offset")
	if err != nil || raw == nil {
		return 0, err
	}

	value, err := readInteger(*raw, "offset")
	if _, tooBig := err.(*strconv.NumError); tooBig {
		return 0, invalid("The value of the `offset` parameter must be an integer.")
	}

	return value, err
}

// Nil without an error means the parameter is absent.
func readSingleValue(params url.Values, name string) (*string, error) {
	values := params[name]

	if len(values) == 0 {
		return nil, nil
	}

	if len(values) > 1 {
		return nil, invalid("The `%s` parameter must be given once.", name)
	}

	return &values[0], nil
}

// Digits only, so "-5" and "+5" are rejected too. A *strconv.NumError means the digits did not fit into an int.
func readInteger(raw, name string) (int, error) {
	if raw == "" || strings.Trim(raw, "0123456789") != "" {
		return 0, invalid("The value of the `%s` parameter must be an integer.", name)
	}

	return strconv.Atoi(raw)
}
