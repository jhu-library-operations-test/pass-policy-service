package rule_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/oa-pass/pass-policy-service/rule"
)

func TestPolicy(t *testing.T) {

	submissionURI := "http://example.org/submission"

	policyJSON := `{
		"description": "Used for unit testing",
		"policy-id": "${submission.foo.policy}",
		"conditions": [
			{
				"endsWith": {
					"good": "${submission.foo.policy}"
				}
			}
		],
		"repositories": [
			{
				"repository-id": "${policy.repository}",
				"selected": true
			},
			{
				"repository-id": "*"
			}
		]
	}`

	policy := rule.Policy{}

	_ = json.Unmarshal([]byte(policyJSON), &policy)

	policies, err := policy.Resolve(&rule.Context{
		SubmissionURI: submissionURI,
		PassClient: testFetcher(map[string]string{
			submissionURI: `{
				"foo": [
					"http://example.org/foo/1",
					"http://example.org/foo/2"
				]
			}`,
			"http://example.org/foo/1": `{
				"policy": "http://example.org/policy/1"
			}`,
			"http://example.org/foo/2": `{
				"policy": "http://example.org/policy/2-good"
			}`,
			"http://example.org/policy/1": `{
				"repository": ["a", "b"]
			}`,
			"http://example.org/policy/2-good": `{
				"repository": ["c", "d"]
			}`,
		}),
	})

	if err != nil {
		t.Fatalf("Failed policy resolve: %+v", err)
	}

	if len(policies) != 1 {
		t.Fatalf("Wrong number of policies: %+v", policies)
	}

	var repos []string
	for _, p := range policies {
		for _, r := range p.Repositories {
			repos = append(repos, r.ID)
		}
	}

	diffs := deep.Equal(repos, []string{"c", "d", "*"})
	if len(diffs) > 0 {
		t.Fatalf("Found differences in expected repositories: %s", strings.Join(diffs, "\n"))
	}

}
