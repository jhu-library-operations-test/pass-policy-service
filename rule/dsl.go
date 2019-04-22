package rule

import (
	"github.com/pkg/errors"
)

// DSL encapsulates to a policy rules document
type DSL struct {
	Schema   string   `json:"$schema"`
	Policies []Policy `json:"policy-rules"`
}

// Resolve parses a policy rules document into a set of concrete rules, with all
// variables interpolated
func Resolve(doc []byte, variables VariablePinner) ([]Policy, error) {
	rules, err := Validate(doc)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid input policy doc")
	}

	var policies []Policy
	for _, policy := range rules.Policies {
		resolved, err := policy.Resolve(variables)
		if err != nil {
			return policies, errors.Wrapf(err, "could not resolve policy rule")
		}
		policies = append(policies, resolved...)
	}

	return policies, nil
}
