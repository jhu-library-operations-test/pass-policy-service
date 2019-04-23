package web

import (
	"net/http"

	"github.com/oa-pass/pass-policy-service/rule"
	"github.com/pkg/errors"
)

type PolicyService struct {
	Rules   rule.PolicyResolver
	Fetcher rule.PassEntityFetcher
}

func NewPolicyService(rulesDoc []byte, fetcher rule.PassEntityFetcher) (service PolicyService, err error) {

	service = PolicyService{Fetcher: fetcher}
	service.Rules, err = rule.Validate(rulesDoc)
	if err != nil {

		return service, errors.Wrapf(err, "could not validate rules dsl")
	}

	return service, nil
}

func (s *PolicyService) RequestPolicies(w http.ResponseWriter, r *http.Request) {

	policyEndpoint := policyEndpoint{s}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		policyEndpoint.handleGet(w, r)
	case http.MethodPost:
		policyEndpoint.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
