package web

import (
	"strings"
)

// BaseURIReplacer replaces public with private BaseURIs and vice versa,
// returning the translated string, as well as a boolean indicating whether
// a baseuri was found and replaced.
type BaseURIReplacer interface {
	PublicWithPrivate(string) (in string, didReplace bool)
	PrivateWithPublic(string) (in string, didReplace bool)
}

// BaseURIs contains public and private BaseURIs, and is a BaseURIReplacer
type BaseURIs struct {
	Public  string
	Private string
}

// PublicWithPrivate replaces he public baseURI of a given string with the
// private baseuri
func (uri BaseURIs) PublicWithPrivate(in string) (string, bool) {
	return uri.doReplace(in, uri.Public, uri.Private)
}

func (uri BaseURIs) PrivateWithPublic(in string) (string, bool) {
	return uri.doReplace(in, uri.Private, uri.Public)
}

func (uri BaseURIs) doReplace(in, a, b string) (string, bool) {
	if !strings.HasPrefix(in, "http") {
		return strings.Join([]string{strings.Trim(b, "/"), strings.TrimLeft(in, "/")}, "/"), true
	}
	if !strings.HasPrefix(in, a) {
		return in, strings.HasPrefix(in, b)
	}

	suffix := strings.TrimLeft(strings.TrimPrefix(in, a), "/")

	return strings.Join([]string{strings.Trim(b, "/"), suffix}, "/"), true
}
