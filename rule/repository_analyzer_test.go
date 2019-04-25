package rule_test

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/oa-pass/pass-policy-service/rule"
)

var emptyRepos = []rule.Repository{}
var emptyRepoList = [][]rule.Repository{}

func TestAnalyzeRequirements(t *testing.T) {

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

func TestElide(t *testing.T) {
	cases := []struct {
		testName     string
		keep         []rule.Repository
		requirements *rule.Requirements
		expected     *rule.Requirements
	}{{
		testName: "keep b from a and (c or d) optional b -> optional b",
		keep: []rule.Repository{
			{ID: "b"},
		},
		requirements: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a"},
			},
			OneOf: [][]rule.Repository{{
				{ID: "c"},
				{ID: "d"},
			}},
			Optional: []rule.Repository{
				{ID: "b"},
			},
		},
		expected: &rule.Requirements{
			Required: emptyRepos,
			OneOf:    emptyRepoList,
			Optional: []rule.Repository{
				{ID: "b"},
			},
		},
	}, {
		testName: "keep {a, b} from c and (a or d) and (b or d) -> optional a, b",
		keep: []rule.Repository{
			{ID: "a"},
			{ID: "b"},
		},
		requirements: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "c"},
			},
			OneOf: [][]rule.Repository{{
				{ID: "a"},
				{ID: "d"},
			}, {
				{ID: "b"},
				{ID: "d"},
			}},
			Optional: emptyRepos,
		},
		expected: &rule.Requirements{
			Required: emptyRepos,
			OneOf:    emptyRepoList,
			Optional: []rule.Repository{
				{ID: "a"},
				{ID: "b"},
			},
		},
	}, {
		testName: "keep {a, b, c} from a and (b or c) optional d -> a and (b or c)",
		keep: []rule.Repository{
			{ID: "a"},
			{ID: "b"},
			{ID: "c"},
		},
		requirements: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a"},
			},
			OneOf: [][]rule.Repository{{
				{ID: "b"},
				{ID: "c"},
			}},
			Optional: []rule.Repository{
				{ID: "d"},
			},
		},
		expected: &rule.Requirements{
			Required: []rule.Repository{
				{ID: "a"},
			},
			OneOf: [][]rule.Repository{{
				{ID: "b"},
				{ID: "c"},
			}},
			Optional: emptyRepos,
		},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.testName, func(t *testing.T) {
			elided := c.requirements.Elide(c.keep)
			diffs := deep.Equal(elided, c.expected)
			if len(diffs) > 0 {
				t.Fatalf("did not get expected results: %s\n%+v", strings.Join(diffs, "\n"), elided)
			}
		})
	}
}
