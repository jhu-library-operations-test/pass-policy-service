package rule_test

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/oa-pass/pass-policy-service/rule"
)

func TestDSL(t *testing.T) {
	submissionURI := "http://example.org/submission"

	json := `{
		"$schema": "https://oa-pass.github.io/pass-policy-service/schemas/policy_config_1.0.json",
		"policy-rules": [
			{
				"description": "Used for unit testing",
				"policy-id": "${submission.foo.policy}",
				"type": "funder",
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
					}
				]
			},
			{
				"description": "Duplicate of first",
				"policy-id": "${submission.foo.policy}",
				"type": "funder",
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
					}
				]
			},
			{
				"description": "Used for unit testing",
				"policy-id": "http://example.org/policy",
				"type": "funder",
				"repositories": [
					{
						"repository-id": "http://example.org/repository",
						"selected": true
					}
				]
			}
		]
	}`

	dsl, err := rule.Validate([]byte(json))
	if err != nil {
		t.Fatalf("rules failed validation %+v", err)
	}

	policies, err := dsl.Resolve(&rule.Context{
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
		t.Fatalf("%+v", err)
	}

	if len(policies) != 2 {
		t.Fatalf("Wrong number of policies %d", len(policies))
	}

	var repos []string
	for _, p := range policies {
		for _, r := range p.Repositories {
			repos = append(repos, r.ID)
		}
	}

	diffs := deep.Equal(repos, []string{"c", "d", "http://example.org/repository"})
	if len(diffs) > 0 {
		t.Fatalf("Found differences in expected repositories: %s", strings.Join(diffs, "\n"))
	}
}
