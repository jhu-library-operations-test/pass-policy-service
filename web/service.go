package web

import (
	"net/http"

	"github.com/oa-pass/pass-policy-service/rule"
	"github.com/pkg/errors"
)

type PolicyService struct {
	Rules   rule.PolicyResolver
	Fetcher rule.PassEntityFetcher
	Replace BaseURIReplacer
}

type requestHandler interface {
	handleGet()
	handlePost()
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
	s.doRequest(&policyRequest{s, r, w}, w, r)
}

func (s *PolicyService) RequestRepositories(w http.ResponseWriter, r *http.Request) {
	s.doRequest(&repositoryRequest{s, r, w}, w, r)
}

func (s *PolicyService) doRequest(handler requestHandler, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		handler.handleGet()
	case http.MethodPost:
		handler.handlePost()
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
