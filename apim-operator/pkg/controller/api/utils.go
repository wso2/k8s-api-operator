package api

import (
	"regexp"
)

// removeVersionTag removes version number in a url provided
func removeVersionTag(url string) string {
	regExpString := `\/v[\d.-]*\/?$`
	regExp := regexp.MustCompile(regExpString)
	return regExp.ReplaceAllString(url, "")
}
