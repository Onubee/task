package service

import (
	"regexp"
	"strconv"
	"strings"
)

func NormalizePrice(priceStr string) (float64, error) {
	re := regexp.MustCompile(`[^0-9.,]`)
	cleaned := re.ReplaceAllString(priceStr, "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ",", ".")
	return strconv.ParseFloat(cleaned, 64)
}
