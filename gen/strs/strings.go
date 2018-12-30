package strs

import (
	"strings"
)

func ToTitle(input string) string {
	var output string
	ss := strings.Split(input, "_")
	for _, s := range ss {
		if s == "" {
			continue
		}
		output += strings.ToUpper(string(s[0])) + s[1:]
	}
	return output
}

func ToTitleNorm(input string) string {
	var output []byte
	var upperCount int
	for i, c := range input {
		switch {
		case c >= 'A' && c <= 'Z':
			if upperCount == 0 || nextIsLower(input, i) {
				output = append(output, byte(c))
			} else {
				output = append(output, byte(c-'A'+'a'))
			}
			upperCount++

		case c >= 'a' && c <= 'z':
			output = append(output, byte(c))
			upperCount = 0

		case c >= '0' && c <= '9':
			if i == 0 {
				panic("common/str: Identifier must start with a character: `" + input + "`")
			}
			output = append(output, byte(c))
			upperCount = 0
		}
	}
	return string(output)
}

func ToSnake(input string) string {
	var output []byte
	var upperCount int
	for i, c := range input {
		switch {
		case c >= 'A' && c <= 'Z':
			if i > 0 && (upperCount == 0 || nextIsLower(input, i)) {
				output = append(output, '_')
			}
			output = append(output, byte(c-'A'+'a'))
			upperCount++

		case c >= 'a' && c <= 'z':
			output = append(output, byte(c))
			upperCount = 0

		case c >= '0' && c <= '9':
			if i == 0 {
				panic("common/str: Identifier must start with a character: `" + input + "`")
			}
			output = append(output, byte(c))

		default:
			panic("common/str: Invalid identifier: `" + input + "`")
		}
	}
	return string(output)
}

func MapToSnake(A []string) []string {
	B := make([]string, len(A))
	for i, s := range A {
		B[i] = ToSnake(s)
	}
	return B
}

// The next character is lower case, but not the last 's'.
//
//     HTMLFile -> html_file
//     URLs     -> urls
func nextIsLower(input string, i int) bool {
	i++
	if i >= len(input) {
		return false
	}
	c := input[i]
	if c == 's' && i == len(input)-1 {
		return false
	}
	return c >= 'a' && c <= 'z'
}

func Abbr(s string) string {
	var res []byte
	for _, c := range ToTitleNorm(s) {
		if c >= 'A' && c <= 'Z' {
			res = append(res, byte(c)-'A'+'a')
		}
	}
	return string(res)
}

func Plural(n int, word, plural string) string {
	if n <= 1 {
		return word
	}
	if plural != "" {
		return plural
	}
	return ToPlural(word)
}

func ToPlural(word string) string {
	l := len(word)
	if l <= 1 {
		return word + "s"
	}
	// Words ending in 's' or 'x'
	l1, l2 := word[l-1], word[l-2]
	if l1 == 's' || l1 == 'x' {
		return word + "es"
	}
	// Words ending in 'ch', 'tsh'
	if l1 == 'h' && (l2 == 's' || l2 == 'c') {
		return word + "es"
	}
	// Words ending in 'o'
	if l1 == 'o' && !isVowel(l2) {
		return word + "es"
	}
	// Words ending in 'y'
	if l1 == 'y' && !isVowel(l2) {
		return word[:l-1] + "ies"
	}
	return word + "s"
}

func isVowel(c byte) bool {
	switch c {
	case 'A', 'E', 'I', 'O', 'U', 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
