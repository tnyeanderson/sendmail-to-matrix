package pkg

import (
	"os"
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
		if FilterMessage(input, skips) != expected {
			t.Fatalf("expected %t result for: %s", expected, input)
		}
	}
}

func TestMessageRenderPrefaceNoSubject(t *testing.T) {
	msg := &Message{
		Preface: "the preface",
		Body:    "the body",
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `the preface
the body`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderPrefaceSubject(t *testing.T) {
	msg := &Message{
		Preface: "the preface",
		Subject: "the subject",
		Body:    "the body",
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `the preface
Subject: the subject
the body`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderEpilogue(t *testing.T) {
	msg := &Message{
		Body:     "the body",
		Epilogue: "the epilogue",
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `the body
the epilogue`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderNonMultipart(t *testing.T) {
	f, _ := os.Open("testdata/mime-non-multipart.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `this is not multipart`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}

}

func TestMessageRenderHTML(t *testing.T) {
	f, _ := os.Open("testdata/mime-html-only.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `this should be sanitized`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}

}

func TestMessageRenderMixedAttachment(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-attachment.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `Subject: Test message
the body`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderAlternativeAttachment(t *testing.T) {
	f, _ := os.Open("testdata/mime-alternative-attachment.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `this is the body text`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderAlternative(t *testing.T) {
	f, _ := os.Open("testdata/mime-alternative-datamotion.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `Subject: This is the subject of a sample message
This is the body text of a sample message.`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderMixedMS(t *testing.T) {
	f, _ := os.Open("testdata/mime-alternative-html.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `this is the body text`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderMixedHtml(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-html.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `this should not be sanitized

example of weird (stupid) proxmox url format:

<http://my.test.url/foo/bar>

this should be sanitized`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}

func TestMessageRenderMixed2(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-2.txt")
	msg, err := NewMessage(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := msg.Render([]byte(DefaultMessageTemplate))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `Subject: Test message from Netscape Communicator 4.7

The Hare and the Tortoise

A HARE one day ridiculed the short feet and slow pace of the Tortoise,
who replied, laughing:  "Though you be swift as the wind, I will beat
you in a race."  The Hare, believing her assertion to be simply
impossible, assented to the proposal; and they agreed that the Fox
should choose the course and fix the goal.  On the day appointed for the
race the two started together.  The Tortoise never for a moment stopped,
but went on with a slow but steady pace straight to the end of the
course.  The Hare, lying down by the wayside, fell fast asleep.  At last
waking up, and moving as fast as he could, he saw the Tortoise had
reached the goal, and was comfortably dozing after her fatigue.

Slow but steady wins the race.

The Farmer and the Stork

A FARMER placed nets on his newly-sown plowlands and caught a
number of Cranes, which came to pick up his seed.  With them he
trapped a Stork that had fractured his leg in the net and was
earnestly beseeching the Farmer to spare his life.  "Pray save
me, Master," he said, "and let me go free this once.  My broken
limb should excite your pity.  Besides, I am no Crane, I am a
Stork, a bird of excellent character; and see how I love and
slave for my father and mother.  Look too, at my feathers--
they are not the least like those of a Crane."   The Farmer
laughed aloud and said, "It may be all as you say, I only know
this:  I have taken you with these robbers, the Cranes, and you
must die in their company."

Birds of a feather flock together.`

	if string(m) != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s\n", m, expected)
	}
}
