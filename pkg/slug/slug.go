package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var slugRegex = regexp.MustCompile(`[^\w-]+`)

// generateSlug normalizes and sanitizes a string to create a URL-friendly slug
func GenerateSlug(t string) string {
	slug := RemoveAccents(t)
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugRegex.ReplaceAllString(slug, "")

	return slug
}

// removeAccents removes diacritical marks (accents) from a string
func RemoveAccents(s string) string {
	/*
		1. Normalize the string to NFC (normalized form decomposed)
		2. Breaks accented letters into two runes: One for the letter and one for the accent
		 	 - Example: "São João"
			 - []rune{'S', 'a', '̃', 'o', ' ', 'J', 'o', '̃', 'a', 'o'}
	*/
	t := norm.NFD.String(s)

	result := make([]rune, 0, len(t))
	for _, r := range t {
		/*
			Mn -> Represents the unicode category "Mark, Nonspacing"
			- Thats include accents, cedillas, umlauts, tildes and any character
				that does not occupy it's own space - that is, combinable accents

			Is() checks if that rune belongs to the given category.
			If it's an accent (Mn), we ignore it with continue.
			If it's a letter or number, we add it to the rune slice.
		*/
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		result = append(result, r)
	}
	return string(result)
}
