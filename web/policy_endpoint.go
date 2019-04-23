package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/oa-pass/pass-policy-service/rule"
	"github.com/pkg/errors"
)

const (
	submissionQueryParam = "submission"
)

type policyEndpoint struct {
	*PolicyService
}

func (p *policyEndpoint) findPolicies(submission string, headers map[string][]string) ([]rule.Policy, error) {
	context := &rule.Context{
		SubmissionURI: submission,
		Headers:       headers,
		PassClient:    p.fetcher,
	}
	policies := make([]rule.Policy, 0, len(p.rules.Policies)*2)
	for _, policy := range p.rules.Policies {
		resolved, err := policy.Resolve(context)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve policies for submission %s", submission)
		}
		policies = append(policies, resolved...)
	}

	return policies, nil
}

func (p *policyEndpoint) sendPolicies(w http.ResponseWriter, r *http.Request, policies []rule.Policy, err error) {
	if err != nil {
		log.Printf("Error resolving policies: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(policies)
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
