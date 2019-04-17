package rule

import (
	"testing"

	"github.com/go-test/deep"
)

func TestIsVariable(t *testing.T) {

	cases := []struct {
		testName     string
		variableName string
		expected     bool
	}{
		{"goodNoDots", "${foo}", true},
		{"goodDots", "${foo.bar.baz}", true},
		{"noBrackets", "$foo.bar.baz", false},
		{"malformed", "${foo.bar.baz}xyz", false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			if IsVariable(c.variableName) != c.expected {
				t.Fatalf("variable status of %s was parsed incorrectly", c.variableName)
			}

			_, ok := toVariable(c.variableName)
			if ok != c.expected {
				t.Fatalf("error converting %s to variable", c.variableName)
			}
		})
	}
}

func TestShift(t *testing.T) {
	v, _ := toVariable("${foo.bar.baz}")
	numSegments := 3

	segments := []variable{{
		segment:     "foo",
		segmentName: "foo",
		fullName:    "foo.bar.baz",
	}, {
		segment:     "bar",
		segmentName: "foo.bar",
		fullName:    "foo.bar.baz",
	}, {
		segment:     "baz",
		segmentName: "foo.bar.baz",
		fullName:    "foo.bar.baz",
	}}

	var i int
	for part, ok := v.shift(); ok; part, ok = part.shift() {

		diffs := deep.Equal(segments[i], part)
		if len(diffs) != 0 {
			t.Errorf("Found differences in variable segment %+v", diffs)
		}

		i++
	}

	if i < numSegments-1 {
		t.Fatalf("Got %d segments, expected %d", i+1, numSegments)
	}
}
