package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/oa-pass/pass-policy-service/rule"
)

const (
	submissionQueryParam = "submission"
)

type policyEndpoint struct {
	*PolicyService
}

// PolicyResult is an item returned in a policy service response, indicating a policy ID, and
// an originating type (funder, institution)
type PolicyResult struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (p *policyEndpoint) findPolicies(submission string, headers map[string][]string) ([]rule.Policy, error) {
	context := &rule.Context{
		SubmissionURI: submission,
		Headers:       headers,
		PassClient:    p.Fetcher,
	}

	return p.Rules.Resolve(context)
}

func (p *policyEndpoint) sendPolicies(w http.ResponseWriter, r *http.Request, policies []rule.Policy, err error) {
	if err != nil {
		log.Printf("Error resolving policies: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []PolicyResult
	for _, policy := range policies {
		results = append(results, PolicyResult{
			ID:   p.replace(policy.ID),
			Type: policy.Type,
		})
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(results)
	if err != nil {
		log.Printf("error encoding JSON response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *policyEndpoint) handleGet(w http.ResponseWriter, r *http.Request) {
	uri, ok := r.URL.Query()[submissionQueryParam]
	if !ok {
		// It would be nice to provide a pretty html page
		http.Error(w, "No submission query param provided", http.StatusBadRequest)
		return
	}

	policies, err := p.findPolicies(uri[0], r.Header)
	p.sendPolicies(w, r, policies, err)
}

func (p *policyEndpoint) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(w,
			"expected media type application/x-www-form-urlencoded, instead got "+
				r.Header.Get("Content-Type"),
			http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form input: "+err.Error(), http.StatusInternalServerError)
		return
	}

	url := r.Form.Get(submissionQueryParam)
	if url == "" {
		http.Error(w, "No submission value provided", http.StatusBadRequest)
		return
	}

	policies, err := p.findPolicies(url, r.Header)
	p.sendPolicies(w, r, policies, err)
}
