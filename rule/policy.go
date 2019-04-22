package rule

import (
	"github.com/pkg/errors"
)

// Policy encapsulates a policy rule.
type Policy struct {
	ID           string       `json:"policy-id"`
	Description  string       `json:"description"`
	Repositories []Repository `json:"repositories"`
	Conditions   []Condition  `json:"conditions"`
}

// resolve interpolates any variables in a policy.  if the policy ID resolves to a list,
// it returns a list of resolved policies, each one with an ID from that list.
func (p Policy) resolve(vaiables VariableResolver) (policies []Policy, err error) {

	var resolvedPolicies []Policy

	// If the policy ID is a variable, we need to resolve/expand it.  If the result is
	// a list of IDs, we return a list of policies, each one with an ID from the list
	if IsVariable(p.ID) {
		resolvedIDs, err := vaiables.Resolve(p.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve property ID %s", p.ID)
		}

		for _, id := range resolvedIDs {

			// Now that we have a concrete ID, resolve any other variables elsewhere in the
			// policy.  Some of them may depend on knowing the ID we just found
			resolved, err := Policy{
				ID:           id,
				Description:  p.Description,
				Repositories: p.Repositories,
				Conditions:   p.Conditions,
			}.resolve(vaiables)

			if err != nil {
				return nil, errors.Wrapf(err, "could not resolve policy rule for %s", id)
			}

			resolvedPolicies = append(resolvedPolicies, resolved...)
		}
	} else {

		// Individual policy.  Resolve the repositories section, and filter by condition to see if
		// it is applicable
		resolvedPolicies = []Policy{p}

		if p.Repositories, err = p.resolveRepositories(vaiables); err != nil {
			return nil, errors.Wrapf(err, "could not resolve repositories in policy %s", p.ID)
		}

		if ok, err := p.applyConditions(vaiables); ok && err != nil {
			resolvedPolicies = append(resolvedPolicies)
		}
	}

	return resolvedPolicies, nil
}

// resolveRepositories replaces any variables in the repository section of a policy.  If repository ID
// is a variable that expands into a list of IDs, then we can have multiple repositories.
func (p Policy) resolveRepositories(variables VariableResolver) ([]Repository, error) {
	var resolved []Repository
	for _, repo := range p.Repositories {
		repos, err := repo.Resolve(variables)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve repositories for %s", p.ID)
		}

		resolved = append(resolved, repos...)
	}
	return resolved, nil
}

// Filter based on evaluating conditions, if there are any
func (p Policy) applyConditions(variables VariableResolver) (bool, error) {
	for _, cond := range p.Conditions {
		if ok, err := cond.Apply(variables); !ok || err != nil {
			return ok, err
		}
	}

	return true, nil
}
