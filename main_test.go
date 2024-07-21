package main

import (
	"os"
	"testing"
)

func TestBuildMessageAlternative(t *testing.T) {
	f, _ := os.Open("testdata/mime-alternative-datamotion.txt")
	m := buildMessage(f, "")
	expected := `Subject: This is the subject of a sample message
This is the body text of a sample message.
`
	if m != expected {
		t.Fail()
	}
}

func TestBuildMessageMixedMS(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-ms.txt")
	m := buildMessage(f, "")
	expected := `this is the body text
`
	if m != expected {
		t.Fail()
	}
}

func TestBuildMessageMixedHtml(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-html.txt")
	m := buildMessage(f, "")
	expected := `this should not be sanitized

example of weird (stupid) proxmox url format:

<http://my.test.url/foo/bar>

this should be sanitized
`
	if m != expected {
		t.Fail()
	}
}

func TestBuildMessageMixed2(t *testing.T) {
	f, _ := os.Open("testdata/mime-mixed-2.txt")
	m := buildMessage(f, "")
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

Birds of a feather flock together.
`
	if m != expected {
		t.Fail()
	}
}
