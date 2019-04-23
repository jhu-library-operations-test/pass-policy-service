package rule_test

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/oa-pass/pass-policy-service/rule"
)

func TestResolveRepository(t *testing.T) {
	cases := []struct {
		testName   string
		resolver   testResolver
		repository rule.Repository
		expected   []rule.Repository
	}{
		{
			testName: "nothingToResolve",
			repository: rule.Repository{
				ID:       "foo",
				Selected: true,
			},
			expected: []rule.Repository{{
				ID:       "foo",
				Selected: true,
			}},
		},
		{
			testName: "oneRepository",
			repository: rule.Repository{
				ID:       "${foo.bar}",
				Selected: true,
			},
			expected: []rule.Repository{{
				ID:       "foo",
				Selected: true,
			}},
			resolver: map[string][]string{
				"${foo.bar}": {"foo"},
			},
		},
		{
			testName: "multiRepository",
			repository: rule.Repository{
				ID:       "${foo.bar}",
				Selected: true,
			},
			expected: []rule.Repository{{
				ID:       "foo",
				Selected: true,
			}, {
				ID:       "bar",
				Selected: true,
			}},
			resolver: map[string][]string{
				"${foo.bar}": {"foo", "bar"},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			resolved, err := c.repository.Resolve(c.resolver)
			if err != nil {
				t.Fatalf("Error resolving repository %+v", err)
			}
			diffs := deep.Equal(resolved, c.expected)
			if len(diffs) > 0 {
				t.Fatalf("resolved repositories are not expected %+v", strings.Join(diffs, "\n"))
			}
		})
	}
}
