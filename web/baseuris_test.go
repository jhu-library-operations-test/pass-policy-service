package web_test

import (
	"testing"

	"github.com/oa-pass/pass-policy-service/web"
)

type testParams struct {
	in, out    string
	didReplace bool
}

func TestPrivateWithPublic(t *testing.T) {
	cases := []struct {
		testName string
		replace  func(string) (string, bool)
		params   testParams
	}{{
		testName: "no match",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar",
		}.PrivateWithPublic,
		params: testParams{
			in:         "http://does-not-match/foo",
			out:        "http://does-not-match/foo",
			didReplace: false,
		},
	}, {
		testName: "no replace",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar",
		}.PrivateWithPublic,
		params: testParams{
			in:         "http://example.org/foo/baz",
			out:        "http://example.org/foo/baz",
			didReplace: true,
		},
	}, {
		testName: "public to private no trailing slashes",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar",
		}.PublicWithPrivate,
		params: testParams{
			in:         "http://example.org/foo/path/to/whatever",
			out:        "http://example.org/bar/path/to/whatever",
			didReplace: true,
		},
	}, {
		testName: "public to private with trailing slashes",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo/",
			Private: "http://example.org/bar/",
		}.PublicWithPrivate,
		params: testParams{
			in:         "http://example.org/foo/path/to/whatever",
			out:        "http://example.org/bar/path/to/whatever",
			didReplace: true,
		},
	}, {
		testName: "public to private with mixed trailing slashes",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo/",
			Private: "http://example.org/bar",
		}.PublicWithPrivate,
		params: testParams{
			in:         "http://example.org/foo/path/to/whatever",
			out:        "http://example.org/bar/path/to/whatever",
			didReplace: true,
		},
	}, {
		testName: "public to private with mixed trailing slashes the other way",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar/",
		}.PublicWithPrivate,
		params: testParams{
			in:         "http://example.org/foo/path/to/whatever",
			out:        "http://example.org/bar/path/to/whatever",
			didReplace: true,
		},
	}, {
		testName: "private to public with mixed trailing slashes",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar/",
		}.PrivateWithPublic,
		params: testParams{
			in:         "http://example.org/bar/path/to/whatever",
			out:        "http://example.org/foo/path/to/whatever",
			didReplace: true,
		},
	}, {
		testName: "relative URI slash",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar/",
		}.PrivateWithPublic,
		params: testParams{
			in:         "/path/to/whatever",
			out:        "http://example.org/foo/path/to/whatever",
			didReplace: true,
		},
	}, {
		testName: "relative URI no slash",
		replace: web.BaseURIs{
			Public:  "http://example.org/foo",
			Private: "http://example.org/bar/",
		}.PrivateWithPublic,
		params: testParams{
			in:         "path/to/whatever",
			out:        "http://example.org/foo/path/to/whatever",
			didReplace: true,
		},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			out, didReplace := c.replace(c.params.in)
			if out != c.params.out {
				t.Fatalf("Expected %s, got %s", c.params.out, out)
			}
			if didReplace != c.params.didReplace {
				t.Fatalf("Expected %t, got %t", c.params.didReplace, didReplace)
			}
		})
	}
}
