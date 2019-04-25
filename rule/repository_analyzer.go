package rule

import (
	"sort"
	"strings"
)

// Requirements encapsulates deposit requirements by sorting repositories
// into "required", 'oneOf', and 'optional' buckets
type Requirements struct {
	Required []Repository
	OneOf    [][]Repository
	Optional []Repository
}

// AnalyzeRequirements analyzes a list of policies, and returns
// repository requirements
func AnalyzeRequirements(policies []Policy) *Requirements {

	// First, sort into "required" and "one of", straight from the policy list
	requirements := categorize(policies)
	var optional []Repository

	// If there are required repos, go through the oneOf lists.
	// If the oneOf list has a required repo, then place remaining into optional
	requirements.OneOf, optional = cutFrom(requirements.OneOf, requirements.Required)
	requirements.Optional = append(requirements.Optional, optional...)

	// If we have any required repos, or more than one OneOf group, go through the remaining lists.
	// if there are any stars, discard and place remaining into optional
	if len(requirements.Required) > 0 || len(requirements.OneOf) > 0 {
		requirements.OneOf, optional = cutFrom(requirements.OneOf, []Repository{{ID: "*"}})
		requirements.Optional = append(requirements.Optional, optional...)
	}

	// If there are any optional in oneOf, remove it from optional
	if len(requirements.OneOf) > 0 {
		optional = optional[:0]
		for _, list := range requirements.OneOf {
			for _, optRepo := range requirements.Optional {
				if !repoListContains(list, optRepo) {
					optional = append(optional, optRepo)
				}
			}
		}
		requirements.Optional = optional
	}

	// If there are any optional in required, remove it from optional
	if len(requirements.Required) > 0 {
		optional = optional[:0]
		for _, optRepo := range requirements.Optional {
			if !repoListContains(requirements.Required, optRepo) {
				optional = append(optional, optRepo)
			}
		}

		requirements.Optional = optional
	}

	// if there are no OneOf or required, promote optional to OneOf
	if len(requirements.OneOf)+len(requirements.Required) == 0 {
		requirements.OneOf = append(requirements.OneOf, requirements.Optional)
		requirements.Optional = nil
	}

	// If there is no required, and only one single-membered oneOf, promote it to required.
	if len(requirements.Required) == 0 && len(requirements.OneOf) == 1 && len(requirements.OneOf[0]) == 1 {
		requirements.Required = append(requirements.Required, requirements.OneOf[0][0])
		requirements.OneOf = nil
	}

	return normalize(requirements)
}

// Sort repos from a set of policies into "required" and "one of buckets"
func categorize(policies []Policy) *Requirements {
	requirements := Requirements{}
	for _, p := range policies {
		if len(p.Repositories) == 1 && p.Repositories[0].ID != "*" {
			requirements.Required = append(requirements.Required, p.Repositories[0])
		} else if len(p.Repositories) > 1 {
			requirements.OneOf = append(requirements.OneOf, p.Repositories)
		}
	}

	return normalize(&requirements)
}

func normalize(in *Requirements) *Requirements {
	in.Required = uniqueRepos(in.Required)
	in.OneOf = uniqueRepoLists(in.OneOf)
	in.Optional = uniqueRepos(in.Optional)

	return in
}

// Sort and make a list of repos unique.  Where two members pointing to the same
// repo differ in their selected value, "true" wins.
func uniqueRepos(repos []Repository) []Repository {
	uniqueRepos := make([]Repository, 0, len(repos))
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].ID < repos[j].ID
	})

	last := &Repository{}
	for _, repo := range repos {
		if repo.ID == last.ID {
			last.Selected = last.Selected || repo.Selected
		} else {
			uniqueRepos = append(uniqueRepos, repo)
			last = &uniqueRepos[len(uniqueRepos)-1]
		}
	}

	return uniqueRepos
}

func repoListKey(repos []Repository) string {
	uris := make([]string, 0, len(repos))
	for _, r := range repos {
		uris = append(uris, r.ID)
	}

	sort.Strings(uris)
	return strings.Join(uris, ";")
}

type keyedRepoList struct {
	key  string
	list []Repository
}

func uniqueRepoLists(lists [][]Repository) [][]Repository {

	uniqueLists := make([][]Repository, 0, len(lists))
	keyedLists := make([]*keyedRepoList, len(lists))
	for i, v := range lists {
		keyedLists[i] = &keyedRepoList{
			key:  repoListKey(v),
			list: uniqueRepos(v),
		}
	}

	sort.Slice(keyedLists, func(i, j int) bool {
		return keyedLists[i].key < keyedLists[j].key
	})

	last := &keyedRepoList{}
	for _, list := range keyedLists {
		if list.key != last.key {
			uniqueLists = append(uniqueLists, list.list)
			last = list
		}
	}

	return uniqueLists
}

// Iterate through a list of lists of repositories.  If any of the lists contains
// a member of the cutlist, remove that list from the list, and return the remaining members.
// For example{{a,b},{c,d}} with a cutlist of {b} would return {{c,d}} and {a}
func cutFrom(lists [][]Repository, cutlist []Repository) (resultingList [][]Repository, remaining []Repository) {
	if len(cutlist) == 0 {
		return lists, remaining
	}

	for _, list := range lists {
		for _, cut := range cutlist {
			if repoListContains(list, cut) {
				for _, member := range list {
					if member.ID != cut.ID {
						remaining = append(remaining, member)
					}
				}
			} else {
				resultingList = append(resultingList, list)
			}
		}
	}
	return
}

func repoListContains(list []Repository, repo Repository) bool {
	for _, member := range list {
		if member.ID == repo.ID {
			return true
		}
	}
	return false
}
