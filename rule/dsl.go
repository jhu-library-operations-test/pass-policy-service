package rule

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// DSL encapsulates to a policy rules document
type DSL struct {
	Schema   string   `json:"$schema"`
	Policies []Policy `json:"policy-rules"`
}

// VariableResolver resolves a variable (a string of form "${foo.bar.baz}"), and returns
// a list of values
type VariableResolver interface {
	Resolve(varString string) ([]string, error)
}

// Resolve parses a policy rules document into a set of concrete rules, with all
// variables interpolated
func Resolve(doc []byte, variables VariableResolver) ([]Policy, error) {
	err := Validate(doc)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid input policy doc")
	}

	rules := &DSL{}
	_ = json.Unmarshal(doc, rules)

	var policies []Policy
	for _, policy := range rules.Policies {
		resolved, err := policy.resolve(variables)
		if err != nil {
			return policies, errors.Wrapf(err, "could not resolve policy rule")
		}
		policies = append(policies, resolved...)
	}

	return policies, nil
}
