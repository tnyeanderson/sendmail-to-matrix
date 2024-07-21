package main

import (
	"testing"

	"github.com/go-test/deep"
)

func TestMergeConfigs(t *testing.T) {
	fromFlags := &Config{
		Server:  "flagserver",
		Token:   "flagtoken",
		Room:    "flagroom",
		Preface: "flagpreface",
		Skips:   []string{"flagskip"},
	}

	fromFile := &Config{
		Server:  "fileserver",
		Token:   "filetoken",
		Room:    "fileroom",
		Preface: "filepreface",
		Skips:   []string{"fileskip"},
	}

	// Empty file config
	t1 := mergeConfigs(*fromFlags, Config{})
	if diff := deep.Equal(t1, fromFlags); diff != nil {
		t.Error(diff)
	}

	// Empty flag config
	t2 := mergeConfigs(Config{}, *fromFile)
	if diff := deep.Equal(t2, fromFile); diff != nil {
		t.Error(diff)
	}

	// Merged
	expected := &Config{
		Server:  "flagserver",
		Token:   "flagtoken",
		Room:    "flagroom",
		Preface: "flagpreface",
		Skips:   []string{"fileskip", "flagskip"},
	}
	t3 := mergeConfigs(*fromFlags, *fromFile)
	if diff := deep.Equal(t3, expected); diff != nil {
		t.Error(diff)
	}
}
