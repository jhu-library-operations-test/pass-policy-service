package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/oa-pass/pass-policy-service/rule"
)

const (
	submissionQueryParam = "submission"
)

type policyRequest struct {
	*PolicyService
	req  *http.Request
	resp http.ResponseWriter
}

// PolicyResult is an item returned in a policy service response, indicating a policy ID, and
// an originating type (funder, institution)
type PolicyResult struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (p *policyRequest) findPolicies(submission string, headers map[string][]string) ([]rule.Policy, error) {
	context := &rule.Context{
		SubmissionURI: submission,
		Headers:       headers,
		PassClient:    p.Fetcher,
	}

	return p.Rules.Resolve(context)
}

func (p *policyRequest) sendPolicies(policies []rule.Policy, err error) {
	if err != nil {
		log.Printf("Error resolving policies: %+v", err)
		http.Error(p.resp, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []PolicyResult
	for _, policy := range policies {
		uri, _ := p.Replace.PrivateWithPublic(policy.ID)
		results = append(results, PolicyResult{
			ID:   uri,
			Type: policy.Type,
		})
	}

	encoder := json.NewEncoder(p.resp)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(results)
	if err != nil {
		log.Printf("error encoding JSON response: %s", err)
		http.Error(p.resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *policyRequest) handleGet() {
	uri, ok := p.req.URL.Query()[submissionQueryParam]
	if !ok {
		// It would be nice to provide a pretty html page
		http.Error(p.resp, "No submission query param provided", http.StatusBadRequest)
		return
	}

	privateSubmissionURI, ok := p.Replace.PublicWithPrivate(uri[0])
	if !ok {
		http.Error(p.resp, fmt.Sprintf("submission URI %s does not have the expected PASS baseURI", uri),
			http.StatusInternalServerError)
		return
	}

	policies, err := p.findPolicies(privateSubmissionURI, p.req.Header)
	p.sendPolicies(policies, err)
}

func (p *policyRequest) handlePost() {
	if p.req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(p.resp,
			"expected media type application/x-www-form-urlencoded, instead got "+
				p.req.Header.Get("Content-Type"),
			http.StatusBadRequest)
		return
	}

	if err := p.req.ParseForm(); err != nil {
		http.Error(p.resp, "Could not parse form input: "+err.Error(), http.StatusInternalServerError)
		return
	}

	url := p.req.Form.Get(submissionQueryParam)
	if url == "" {
		http.Error(p.resp, "No submission value provided", http.StatusBadRequest)
		return
	}

	policies, err := p.findPolicies(url, p.req.Header)
	p.sendPolicies(policies, err)
}
