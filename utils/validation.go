package utils

import (
	"regexp"
	"time"
)

func IsValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
func IsValidTime(timeStr string) bool {
	_, err := time.Parse("15:04", timeStr)
	return err == nil
}

func IsValidPrice(price string) bool {
	match, _ := regexp.MatchString(`^\d+\.\d{2}$`, price)
	return match
}

func IsValidRetailer(retailer string) bool {
	match, _ := regexp.MatchString(`^[\w\s\-&]+$`, retailer)
	return match
}