package rule

import (
	"github.com/pkg/errors"
)

type Repository struct {
	ID       string `json:"repository-id"`
	Selected bool   `json:"selected"`
}

func (r Repository) Resolve(variables VariableResolver) ([]Repository, error) {
	var resolvedRepositories []Repository

	if IsVariable(r.ID) {
		resolvedIDs, err := variables.Resolve(r.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve property ID %s", r.ID)
		}

		for _, id := range resolvedIDs {
			resolvedRepositories = append(resolvedRepositories,
				Repository{
					ID:       id,
					Selected: r.Selected,
				},
			)
		}
	} else {
		resolvedRepositories = []Repository{r}
	}

	return resolvedRepositories, nil

}
