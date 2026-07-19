package command

import (
	"fmt"
	"net/url"
	"strconv"
)

func (d *DownloadCommand) addPaginationParams(baseURL string, page int, limit int) string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" {
		return fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page, limit)
	}

	query := parsedURL.Query()
	query.Set("page", strconv.Itoa(page))
	query.Set("limit", strconv.Itoa(limit))
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}
