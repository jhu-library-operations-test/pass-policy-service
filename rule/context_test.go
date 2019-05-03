package rule_test

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/oa-pass/pass-policy-service/rule"
)

func TestContextResolve(t *testing.T) {
	submissionURI := "http://example.org/submissionUri"

	cases := []struct {
		testName      string
		fetcher       testFetcher
		headers       map[string][]string
		varName       string
		expectedValue []string
	}{{
		testName:      "notVariable",
		varName:       "$moo",
		expectedValue: []string{"$moo"},
	}, {
		testName: "noValue",
		varName:  "${submission.baz}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo":"bar",
				"not":"used"
			}`,
		},
		expectedValue: []string{},
	}, {
		testName: "noValueMulti",
		varName:  "${submission.baz.foo}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo":"bar",
				"not":"used"
			}`,
		},
		expectedValue: []string{},
	}, {
		testName:      "submission",
		varName:       "${submission}",
		expectedValue: []string{submissionURI},
	}, {
		testName: "submissionProperty",
		varName:  "${submission.foo}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo":"bar",
				"not":"used"
			}`,
		},
		expectedValue: []string{"bar"},
	}, {
		testName: "headerProperty",
		varName:  "${header.foo}",
		headers: map[string][]string{
			"foo": {"bar"},
		},
		expectedValue: []string{"bar"},
	}, {
		testName: "jsonStringProperty",
		varName:  "${submission.foo.bar}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo":"{\"bar\":\"baz\"}"
			}`,
		},
		expectedValue: []string{"baz"},
	}, {
		testName: "traverseSingleObjects",
		varName:  "${submission.foo.bar}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo": "https:/example.org/foo/1" 
			}`,
			"https:/example.org/foo/1": `{
				"bar": "http://example.org/baz"
			}`,
		},
		expectedValue: []string{"http://example.org/baz"},
	}, {
		testName: "traverseSingleObjectsMultiHop",
		varName:  "${submission.foo.bar.baz}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo": "https:/example.org/foo/1" 
			}`,
			"https:/example.org/foo/1": `{
				"bar": "http://example.org/baz"
			}`,
			"http://example.org/baz": `{
				"baz": "http://example.org/awesome"
			}`,
		},
		expectedValue: []string{"http://example.org/awesome"},
	}, {
		testName: "traverseSingleListObjects",
		varName:  "${submission.foo.bar}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo": ["http:/example.org/singleGrant"] 
			}`,
			"http:/example.org/singleGrant": `{
				"bar": "http://example.org/directFunder"
			}`,
		},
		expectedValue: []string{"http://example.org/directFunder"},
	}, {
		testName: "traverseManyListObjects",
		varName:  "${submission.foo.bar.rhubarb}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo": [
					"http:/example.org/foo/1",
					"http:/example.org/foo/2"
				] 
			}`,
			"http:/example.org/foo/1": `{
				"bar": [
					"http:/example.org/bar/1",
					"http:/example.org/bar/2"
				]
			}`,
			"http:/example.org/foo/2": `{
				"bar": [
					"http:/example.org/bar/1",
					"http:/example.org/bar/3"
				]
			}`,
			"http:/example.org/bar/1": `{
				"rhubarb": [
					"a",
					"b"
				]
			}`,
			"http:/example.org/bar/2": `{
				"rhubarb": [
					"b",
					"c"
				]
			}`,
			"http:/example.org/bar/3": `{
				"rhubarb": [
					"d",
					"e"
				]
			}`,
		},
		expectedValue: []string{"a", "b", "c", "d", "e"},
	}, {
		testName: "traverse many ist objects no results",
		varName:  "${submission.foo.bar.rhubarb}",
		fetcher: map[string]string{
			submissionURI: `{
				"foo": [
					"http:/example.org/foo/1",
					"http:/example.org/foo/2"
				] 
			}`,
			"http:/example.org/foo/1": `{
				"bar": [
					"http:/example.org/bar/1",
					"http:/example.org/bar/2"
				]
			}`,
			"http:/example.org/foo/2": `{
				"bar": [
					"http:/example.org/bar/1",
					"http:/example.org/bar/3"
				]
			}`,
			"http:/example.org/bar/1": `{
				"nope": [
					"a",
					"b"
				]
			}`,
			"http:/example.org/bar/2": `{
				"nope": [
					"b",
					"c"
				]
			}`,
			"http:/example.org/bar/3": `{
				"nope": [
					"d",
					"e"
				]
			}`,
		},
		expectedValue: []string{},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			cxt := rule.Context{
				SubmissionURI: submissionURI,
				PassClient:    c.fetcher,
				Headers:       c.headers,
			}

			vals, err := cxt.Resolve(c.varName)
			if err != nil {
				t.Fatalf("Error resolving variable %s: %+v", c.varName, err)
			}

			diffs := deep.Equal(vals, c.expectedValue)
			if len(diffs) != 0 {
				t.Fatalf("Found differences in expected values: %s", strings.Join(diffs, "\n"))
			}

		})
	}
}

// Use the same context for multiple variable resolutions
func TestContextMultipleResolve(t *testing.T) {

	submissionURI := "http://example.org/submission"

	varNames := []string{
		"${submission.foo.bar.rhubarb}",
		"${submission.foo}",
		"${submission.foo.moo}",
		"${submission.foo.bar.rhubarb}",
		"${submission.single.value}",
		"${submission.single}",
	}

	expectedValues := []interface{}{
		[]string{"a", "b", "c", "d", "e"},
		[]string{
			"http:/example.org/foo/1",
			"http:/example.org/foo/2",
		},
		[]string{"cow"},
		[]string{"a", "b", "c", "d", "e"},
		[]string{"moo"},
		[]string{"http://example.org/single"},
	}

	fetcher := testFetcher(map[string]string{
		submissionURI: `{
				"foo": [
					"http:/example.org/foo/1",
					"http:/example.org/foo/2"
				],
				"single": "http://example.org/single"
			}`,
		"http://example.org/single": `{
			"value": "moo"
		}`,
		"http:/example.org/foo/1": `{
				"bar": [
					"http:/example.org/bar/1",
					"http:/example.org/bar/2"
				], 
				"moo": "cow"
			}`,
		"http:/example.org/foo/2": `{
				"bar": [
					"http:/example.org/bar/1",
					"http:/example.org/bar/3"
				]
			}`,
		"http:/example.org/bar/1": `{
				"rhubarb": [
					"a",
					"b"
				]
			}`,
		"http:/example.org/bar/2": `{
				"rhubarb": [
					"b",
					"c"
				]
			}`,
		"http:/example.org/bar/3": `{
				"rhubarb": [
					"d",
					"e"
				]
			}`,
	})

	cxt := rule.Context{
		SubmissionURI: submissionURI,
		PassClient:    fetcher,
	}

	for i, v := range varNames {
		v := v
		i := i
		t.Run(v, func(t *testing.T) {

			vals, err := cxt.Resolve(v)
			if err != nil {
				t.Fatalf("Error resolving variable %s: %+v", v, err)
			}

			diffs := deep.Equal(vals, expectedValues[i])
			if len(diffs) != 0 {
				t.Fatalf("Found differences in expected values: %s", strings.Join(diffs, "\n"))
			}
		})
	}
}

func TestContextErrors(t *testing.T) {
	submissionURI := "http://example.org/submission"

	cases := []struct {
		testName string
		fetcher  rule.PassEntityFetcher
		headers  map[string][]string
		varName  string
	}{{
		testName: "bad ubmission URI",
		fetcher:  testFetcher(nil),
		varName:  "${submission.foo}",
	}, {
		testName: "not a URI",
		varName:  "${submission.foo.bar}",
		fetcher: testFetcher(map[string]string{
			submissionURI: `{
				"foo": ["moo"] 
			}`,
		}),
	}, {
		testName: "not a String",
		varName:  "${submission.foo}",
		fetcher: testFetcher(map[string]string{
			submissionURI: `{
					"foo": {"bar": "baz"}
				}`,
		}),
	}, {
		testName: "heterogeneous List",
		varName:  "${submission.foo.bar}",
		fetcher: testFetcher(map[string]string{
			submissionURI: `{
					"foo": [
						"http://example.org/whatever",
						14.8
					]
				}`,
		}),
	}, {
		testName: "bad JSON",
		varName:  "${submission.foo.bar.baz}",
		fetcher: testFetcher(map[string]string{
			submissionURI: `{
					"foo": [
						"http://example.org/bar"
					
				}`,
			"http://example.org/bar": `{
				"baz": "hello"
				"foo": "bar"
			}`,
		}),
	}, {
		testName: "does not resolve",
		varName:  "${submission.foo.bar}",
		fetcher:  errFetcher{},
	}, {
		testName: "unexpected type",
		varName:  "${submission.foo.bar}",
		fetcher: testFetcher(map[string]string{
			submissionURI: `{
					"foo": 7.4
				}`,
		}),
	}}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			cxt := rule.Context{
				SubmissionURI: submissionURI,
				PassClient:    c.fetcher,
				Headers:       c.headers,
			}

			_, err := cxt.Resolve(c.varName)
			if err == nil {
				t.Fatalf("Should have exited with an error")
			}
		})
	}
}
