package strs_test

import (
	"testing"

	. "github.com/ng-vu/sqlgen/gen/strs"
)

func TestToTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"hello_world", "HelloWorld"},
		{"t3pl", "T3pl"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToTitle(tt.input); got != tt.want {
				t.Errorf("toTitle(\"%v\") = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToTitleNorm(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HelloWorld", "HelloWorld"},
		{"T3PL", "T3Pl"},
		{"AccountID", "AccountId"},
		{"AccountIDs", "AccountIds"},
		{"ImageURLs", "ImageUrls"},
		{"HTML", "Html"},
		{"HTMLs", "Htmls"},
		{"HTMLFile", "HtmlFile"},
		{"HTMLFileURL", "HtmlFileUrl"},
		{"HTMLFileURLs", "HtmlFileUrls"},
		{"GetHTMLFile", "GetHtmlFile"},
		{"GetHTMLFileURL", "GetHtmlFileUrl"},
		{"PString", "PString"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToTitleNorm(tt.input); got != tt.want {
				t.Errorf("toTitleNorm(\"%v\") = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HelloWorld", "hello_world"},
		{"T3PL", "t3pl"},
		{"AccountID", "account_id"},
		{"AccountIDs", "account_ids"},
		{"HTML", "html"},
		{"HTMLs", "htmls"},
		{"HTMLFile", "html_file"},
		{"HTMLFileURL", "html_file_url"},
		{"HTMLFileURLs", "html_file_urls"},
		{"GetHTMLFile", "get_html_file"},
		{"GetHTMLFileURL", "get_html_file_url"},
		{"PString", "p_string"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToSnake(tt.input); got != tt.want {
				t.Errorf("toSnake(\"%v\") = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToPlural(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Single letter
		{"A", "As"},
		{"a", "as"},
		{"O", "Os"},
		{"o", "os"},

		// All caps
		{"HTML", "HTMLs"},
		{"URL", "URLs"},
		{"DO", "DOs"},
		{"DOS", "DOSs"},

		// Title
		{"Do", "Does"},
		{"Directory", "Directories"},

		// s
		{"file", "files"},
		{"user", "users"},

		// s-es | x-es
		{"class", "classes"},
		{"box", "boxes"},

		// ch-es
		{"lunch", "lunches"},
		{"watch", "watches"},

		// sh-es
		{"finish", "finishes"},

		// o-es
		{"do", "does"},
		{"potato", "potatoes"},

		// y-ies
		{"directory", "directories"},
		{"toy", "toys"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToPlural(tt.input); got != tt.want {
				t.Errorf("toPlural(\"%v\") = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
