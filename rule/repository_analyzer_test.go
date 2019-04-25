package rule_test

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/oa-pass/pass-policy-service/rule"
)

func TestAnalyzeRequirements(t *testing.T) {
	emptyRepos := []rule.Repository{}
	emptyRepoList := [][]rule.Repository{}

	cases := []struct {
		testName string
		policies []rule.Policy
		expected *rule.Requirements
	}{{
		testName: "a and (b or c) -> a and (b or c)",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "b", Selected: true},
				{ID: "c", Selected: false},
			},
		}},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a", Selected: true},
			},
			OneOf: [][]rule.Repository{{
				{ID: "b", Selected: true},
				{ID: "c", Selected: false},
			}},
			Optional: emptyRepos,
		},
	}, {
		testName: "a and (b or *) -> a, optional b",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{{ID: "a", Selected: true}},
		}, {
			Repositories: []rule.Repository{
				{ID: "b", Selected: true},
				{ID: "*", Selected: false},
			},
		}},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a", Selected: true},
			},
			OneOf: emptyRepoList,
			Optional: []rule.Repository{
				{ID: "b", Selected: true},
			},
		},
	}, {
		testName: "(a or *) -> a",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
				{ID: "*", Selected: true},
			},
		}},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a", Selected: true},
			},
			OneOf:    emptyRepoList,
			Optional: emptyRepos,
		},
	}, {
		testName: "(a or b) and (b or c) -> (a or b) and (b or c)",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
				{ID: "b", Selected: false},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "b", Selected: true},
				{ID: "c", Selected: false},
			},
		}},
		expected: &rule.Requirements{
			Required: emptyRepos,
			OneOf: [][]rule.Repository{{
				{ID: "a", Selected: true},
				{ID: "b", Selected: false},
			}, {
				{ID: "b", Selected: true},
				{ID: "c", Selected: false},
			}},
			Optional: emptyRepos,
		},
	}, {
		testName: "a and (a or b) and (b or c) -> a and (b or c)",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
				{ID: "b", Selected: false},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "b", Selected: true},
				{ID: "c", Selected: false},
			},
		}},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a", Selected: true},
			},
			OneOf: [][]rule.Repository{{
				{ID: "b", Selected: true},
				{ID: "c", Selected: false},
			}},
			Optional: emptyRepos,
		},
	}, {
		testName: "a and (a or b) -> a optional b",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "a", Selected: false},
				{ID: "b", Selected: true},
			},
		}},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a", Selected: true},
			},
			OneOf: emptyRepoList,
			Optional: []rule.Repository{
				{ID: "b", Selected: true},
			},
		},
	}, {
		testName: "(a or *) and (b or *) -> (a or b)",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
				{ID: "*", Selected: true},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "b", Selected: true},
				{ID: "*", Selected: false},
			},
		}},
		expected: &rule.Requirements{
			Required: emptyRepos,
			OneOf: [][]rule.Repository{{
				{ID: "a", Selected: true},
				{ID: "b", Selected: true},
			}},
			Optional: emptyRepos,
		},
	}, {
		testName: "a and b and (a or b) -> a and b",
		policies: []rule.Policy{{
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "b", Selected: true},
			},
		}, {
			Repositories: []rule.Repository{
				{ID: "a", Selected: true},
				{ID: "b", Selected: true},
			},
		}},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a", Selected: true},
				{ID: "b", Selected: true},
			},
			OneOf:    emptyRepoList,
			Optional: emptyRepos,
		},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			analyzed := rule.AnalyzeRequirements(c.policies)
			diffs := deep.Equal(analyzed, c.expected)
			if len(diffs) > 0 {
				t.Fatalf("did not get expected results: %s\n%+v", strings.Join(diffs, "\n"), analyzed)
			}
		})
	}
}
