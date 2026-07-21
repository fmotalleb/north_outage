package collector

import (
	"regexp"
	"strings"
)

var replacements = map[rune]rune{
	// Kaf
	'ك': 'ک',

	// Yeh & Alef Maksura
	'ي': 'ی',
	'ى': 'ی',
	'ئ': 'ی',

	// Heh variants
	'ة': 'ه',
	'ۀ': 'ه',
	'ہ': 'ه',
	'ھ': 'ه',

	// Waw variants
	'ؤ': 'و',

	// Digits (Arabic → Persian)
	'٠': '0',
	'١': '1',
	'٢': '2',
	'٣': '3',
	'٤': '4',
	'٥': '5',
	'٦': '6',
	'٧': '7',
	'٨': '8',
	'٩': '9',

	// Misc
	'-': ' ',
}
var sanitizer = regexp.MustCompile(`\s\s+`)

func persianFixer(input string) string {
	var b strings.Builder
	for _, r := range input {
		// Remove Arabic diacritics
		if r >= 0x064B && r <= 0x0652 {
			continue
		}
		if rep, ok := replacements[r]; ok {
			b.WriteRune(rep)
		} else {
			b.WriteRune(r)
		}
	}
	spaced := b.String()
	result := sanitizer.ReplaceAllString(spaced, " ")
	return result
}
