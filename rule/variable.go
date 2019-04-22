package rule

import (
	"strings"
)

// variable encodes a variable for interpolation, eg. ${foo.bar.baz}, or a
// segment of one, e.g. ${foo.bar} of ${foo.bar.baz}
type variable struct {
	segment     string // e.g. "bar" from ${foo.bar.baz}
	segmentName string // e.g. "foo.bar" from ${foo.bar.baz}
	fullName    string // e.g. "foo.bar.baz" from ${foo.bar.baz}
}

type passThroughResolver struct{}

func (passThroughResolver) Resolve(varString string) ([]string, error) {
	return []string{varString}, nil
}

// IsVariable determines if a string is a variable (e.g. of the form '${foo.bar.baz}')
func IsVariable(text string) bool {
	return strings.HasPrefix(text, "${") &&
		strings.HasSuffix(text, "}")
}

// toVariable creates a Variable from a string like '${foo.bar.baz}' (dollar sign and braces required)
func toVariable(text string) (variable, bool) {
	if !IsVariable(text) {
		return variable{}, false
	}

	return variable{
		fullName: strings.TrimRight(strings.TrimLeft(text, "${"), "}"),
	}, true
}

// shift is used for producing a segment of a variable, e.g. Shift() of ${foo.bar.baz},
// is ${foo}.  Shift() of that ${foo} is ${foo.bar}, and Shift() of ${foo.bar} is ${foo.bar.baz}
func (v variable) shift() (variable, bool) {
	remaining := strings.TrimLeft(strings.TrimPrefix(v.fullName, v.segmentName), ".")

	// no more variable segments
	if remaining == "" {
		return v, false
	}

	shifted := variable{
		fullName: v.fullName,
	}

	if v.segment == "" {
		shifted.segment = strings.Split(v.fullName, ".")[0]
		shifted.segmentName = shifted.segment
	} else {
		parts := strings.Split(remaining, ".")
		shifted.segment = parts[0]
		shifted.segmentName = strings.Join([]string{v.segmentName, parts[0]}, ".")
	}

	return shifted, true
}

func (v variable) prev() variable {
	prev := variable{
		fullName: v.fullName,
	}
	if v.segment == "" {
		return prev
	}

	prev.segmentName = strings.Trim(strings.TrimSuffix(v.segmentName, v.segment), ".")
	segments := strings.Split(prev.segmentName, ".")
	prev.segment = segments[len(segments)-1]

	return prev
}
