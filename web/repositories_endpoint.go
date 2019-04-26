package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/oa-pass/pass-policy-service/rule"
	"github.com/pkg/errors"
)

type repositoryRequest struct {
	*PolicyService
	req  *http.Request
	resp http.ResponseWriter
}

func (re *repositoryRequest) handleGet() {
	uri, ok := re.req.URL.Query()[submissionQueryParam]
	if !ok {
		// It would be nice to provide a pretty html page
		http.Error(re.resp, "No submission query param provided", http.StatusBadRequest)
		return
	}

	re.performRequest(uri[0])
}

func (re *repositoryRequest) handlePost() {
	if re.req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(re.resp,
			"expected media type application/x-www-form-urlencoded, instead got "+
				re.req.Header.Get("Content-Type"),
			http.StatusBadRequest)
		return
	}

	if err := re.req.ParseForm(); err != nil {
		http.Error(re.resp, "Could not parse form input: "+err.Error(), http.StatusInternalServerError)
		return
	}

	url := re.req.Form.Get(submissionQueryParam)
	if url == "" {
		http.Error(re.resp, "No submission value provided", http.StatusBadRequest)
		return
	}

	re.performRequest(url)

}

func (re *repositoryRequest) performRequest(publicSubmissionURI string) {

	// Get the submission URI on the private net
	privateSubmissionURI, ok := re.Replace.PublicWithPrivate(publicSubmissionURI)
	if !ok {
		http.Error(re.resp, fmt.Sprintf("submission URI %s does not have the expected PASS baseURI", privateSubmissionURI),
			http.StatusInternalServerError)
		return
	}

	// Resolve the policies inherently implied by the submission
	fmt.Println("Resolving policies for " + privateSubmissionURI)
	policies, err := re.Rules.Resolve(&rule.Context{
		SubmissionURI: privateSubmissionURI,
		Headers:       re.req.Header,
		PassClient:    re.Fetcher,
	})
	if err != nil {
		log.Printf("Error resolving policies: %+v", err)
		http.Error(re.resp, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the repositories in common with between the "repositories implied by the policies
	// inherent to the submission" vs "repositories of policies listed in effectivePolicies"
	// These are the repositories PASS may need to deposit into.
	privateRepoUrisToDepositInto, err := re.reconcileRepositories(privateSubmissionURI, policies)
	if err != nil {
		log.Printf("Error reconciling policies: %+v", err)
		http.Error(re.resp, err.Error(), http.StatusInternalServerError)
		return
	}

	requirements := rule.AnalyzeRequirements(policies).
		TranslateURIs(re.Replace.PublicWithPrivate). // needed because polcies (from) may contain relative or public URIs
		Keep(privateRepoUrisToDepositInto).
		TranslateURIs(re.Replace.PrivateWithPublic)

	encoder := json.NewEncoder(re.resp)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(requirements)
	if err != nil {
		log.Printf("error encoding JSON response: %s", err)
		http.Error(re.resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

type SubmissionEffectivePolicies struct {
	PolicyURIs []string `json:"effectivePolicies"`
}

// reconcileRepositories compares all repositories implied by the policies enumerated in a
// submission's effectivePolicies, compares it to the list of repositories enumerated by the
// given policy list, and returns the list of repositories in common between the two
func (re *repositoryRequest) reconcileRepositories(
	submission string, policies []rule.Policy) ([]rule.Repository, error) {

	// first, fetch effective policies from the given submission.
	policyData := SubmissionEffectivePolicies{}
	err := re.Fetcher.FetchEntity(submission, &policyData)
	if err != nil {
		return nil, errors.Wrapf(err, "Error retrieving effective policies from submission %s", submission)
	}

	// Then build a map of known policies
	knownPolicies := make(map[string]*rule.Policy, len(policies))
	for i := range policies {
		uri, _ := re.Replace.PublicWithPrivate(policies[i].ID)
		knownPolicies[uri] = &policies[i]
	}

	// For each effective policy from the submission, match it with a known policy, and collect the repositories
	var commonRepositories []rule.Repository
	encounteredRepositories := make(map[string]bool, len(policies))
	for _, effectivePolicy := range policyData.PolicyURIs {
		effectivePolicyURI, ok := re.Replace.PublicWithPrivate(effectivePolicy)
		if !ok {
			return nil, errors.Errorf("policy URI %s does not start with a public or private PASS baseuri", effectivePolicy)
		}

		commonPolicy, ok := knownPolicies[effectivePolicyURI]
		if !ok {
			return nil, errors.Errorf("effective policy %s is not in the list of computed policies: %+v",
				effectivePolicyURI,
				knownPolicies)
		}

		for _, repo := range commonPolicy.Repositories {
			if repo.ID == "*" {
				continue
			}

			repoID, _ := re.Replace.PublicWithPrivate(repo.ID)
			if !encounteredRepositories[repoID] {
				commonRepositories = append(commonRepositories, rule.Repository{ID: repoID})
				encounteredRepositories[repoID] = true
			}
		}
	}

	return commonRepositories, nil
}
