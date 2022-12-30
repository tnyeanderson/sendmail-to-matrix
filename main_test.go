package main

import (
	"os"
	"testing"
)

func TestBuildMessageAlternative(t *testing.T) {
	f, _ := os.Open("testdata/mime-alternative-datamotion.txt")
	m := buildMessage(f)
	expected := `Subject: This is the subject of a sample message
This is the body text of a sample message.
`
	if m != expected {
		t.Fail()
	}
}

func TestBuildMessageMixedMS(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-ms.txt")
	m := buildMessage(f)
	expected := `this is the body text
`
	if m != expected {
		t.Fail()
	}
}
