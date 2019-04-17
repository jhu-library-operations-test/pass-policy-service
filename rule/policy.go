package rule

import (
	"github.com/pkg/errors"
)

type Policy struct {
	ID           string       `json:"policy-id"`
	Description  string       `json:"description"`
	Repositories []Repository `json:"repositories"`
	Conditions   []Condition  `json:"conditions"`
}

func (p Policy) resolve(vaiables VariableResolver) (policies []Policy, err error) {

	var resolvedPolicies []Policy

	if IsVariable(p.ID) {
		resolvedIDs, err := vaiables.Resolve(p.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve property ID %s", p.ID)
		}

		for _, id := range resolvedIDs {

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

func (p Policy) applyConditions(variables VariableResolver) (bool, error) {
	for _, cond := range p.Conditions {
		if ok, err := cond.apply(variables); !ok || err != nil {
			return ok, err
		}
	}

	return true, nil
}
