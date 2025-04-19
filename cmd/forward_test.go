package cmd

import (
	"regexp"
	"testing"
)

func TestFilterMessage(t *testing.T) {
	skips := []*regexp.Regexp{
		regexp.MustCompile("shouldmatch"),
		regexp.MustCompile("wontmatch"),
	}
	tests := map[string]bool{
		"nomatch":                               true,
		"i shouldmatch, because i should match": false,
		"i won't actually match":                true,
	}

	for input, expected := range tests {
		if filterMessage(input, skips) != expected {
			t.Fatalf("expected %t result for: %s", expected, input)
		}
	}
}
