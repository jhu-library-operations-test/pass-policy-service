package rule

import (
	"github.com/pkg/errors"
)

// DSL encapsulates to a policy rules document
type DSL struct {
	Schema   string   `json:"$schema"`
	Policies []Policy `json:"policy-rules"`
}

type PolicyResolver interface {
	Resolve(variables VariablePinner) ([]Policy, error)
}

func (d *DSL) Resolve(variables VariablePinner) ([]Policy, error) {
	var policies []Policy
	for _, policy := range d.Policies {
		resolved, err := policy.Resolve(variables)
		if err != nil {
			return policies, errors.Wrapf(err, "could not resolve policy rule")
		}
		policies = append(policies, resolved...)
	}

	return uniquePolicies(policies), nil
}
