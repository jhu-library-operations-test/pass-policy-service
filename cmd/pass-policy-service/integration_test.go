// +build integration

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/oa-pass/pass-policy-service/rule"
)

const defaultFedoraBaseuri = "http://localhost:8080/fcrepo/rest"

func TestFedoraIntegration(t *testing.T) {
	client := &http.Client{}

	fedora := resourceHelper{t, client}

	nihRepo := fedora.repository("repositories/nih")
	decRepo := fedora.repository("repositories/dec")

	nihPolicy := fedora.policy("policies/nih", []string{nihRepo})
	decPolcy := fedora.policy("policies/dec", []string{decRepo})

	nihFunder := fedora.funder("funders/nih", nihPolicy)
	decFunder := fedora.funder("funders/dec", decPolcy)

	nihGrant := fedora.grant("grants/nih", nihFunder, "")
	decGrant := fedora.grant("grants/dec", "", decFunder)

	submission := fedora.submission("submissions/foo", []string{nihGrant, decGrant})

	get, _ := http.NewRequest(http.MethodGet, policyServiceURI()+"/policies?submission="+submission, nil)
	get.Header.Set("Eppn", "someone@johnshopkins.edu")

	resp, err := client.Do(get)
	if err != nil {
		t.Fatalf("GET request to %s failed: %s", get.RequestURI, err)
	}

	if resp.StatusCode > 299 {
		t.Fatalf("Policy service returned an error on GET to %s: %d", get.URL, resp.StatusCode)
	}

	// Read in the body
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var policies []rule.Policy

	_ = json.Unmarshal(body, &policies)
	if len(policies) != 3 {
		t.Fatalf("Wrong number of policies, got %d", len(policies))
	}
}

func authz(r *http.Request) {
	username, ok := os.LookupEnv("PASS_FEDORA_USER")
	if !ok {
		username = "fedoraAdmin"
	}

	passwd, ok := os.LookupEnv("PASS_FEDORA_PASSWORD")
	if !ok {
		passwd = "moo"
	}

	r.SetBasicAuth(username, passwd)
}

func policyServiceURI() string {
	port, ok := os.LookupEnv("POLICY_SERVICE_PORT")
	if !ok {
		port = "8088"
	}

	host, ok := os.LookupEnv("POLICY_SERVICE_HOST")
	if !ok {
		host = "localhost"
	}

	return fmt.Sprintf("http://%s:%s", host, port)
}

func fedoraURI(uripath string) string {
	return fmt.Sprintf("%s/%s", fedoraBaseURI(), uripath)
}

func fedoraBaseURI() string {
	baseuri, ok := os.LookupEnv("PASS_EXTERNAL_FEDORA_BASEURL")
	if !ok {
		return defaultFedoraBaseuri
	}

	return strings.Trim(baseuri, "/")
}

type resourceHelper struct {
	t *testing.T
	c *http.Client
}

func (r *resourceHelper) submission(path string, funders []string) string {
	uri := fedoraURI(path)

	submission := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.4.jsonld",
		"@id" : "%s",
		"grants": [
			%s
		],
		"@type" : "Submission"
	  }
	  `, uri, jsonList(funders))

	r.putResource(uri, submission)
	return uri
}

func (r *resourceHelper) grant(path, priFunder, dirFunder string) string {
	var funders string
	if priFunder != "" {
		funders = fmt.Sprintf(`"primaryFunder": "%s",`, priFunder) + "\n"
	}
	if dirFunder != "" {
		funders = fmt.Sprintf(`%s"directFunder": "%s",`, funders, dirFunder) + "\n"
	}

	uri := fedoraURI(path)

	funder := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.4.jsonld",
		"@id" : "%s",
		 %s
		"@type" : "Grant"
	  }
	  `, uri, funders)

	r.putResource(uri, funder)
	return uri

}

func (r *resourceHelper) funder(path, policy string) string {
	uri := fedoraURI(path)

	funder := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.4.jsonld",
		"@id" : "%s",
		"policy": "%s",
		"@type" : "Funder"
	  }
	  `, uri, policy)

	r.putResource(uri, funder)
	return uri
}

func (r *resourceHelper) policy(path string, repositories []string) string {
	uri := fedoraURI(path)

	policy := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.4.jsonld",
		"@id" : "%s",
		"repositories": [
			%s
		],
		"@type" : "Policy"
	  }
	  `, uri, jsonList(repositories))

	r.putResource(uri, policy)
	return uri
}

func jsonList(list []string) string {
	var jsonList []string
	for _, item := range list {
		jsonList = append(jsonList, fmt.Sprintf(`"%s"`, item))
	}

	return strings.Trim(strings.Join(jsonList, ",\n"), ",\n")
}

func (r *resourceHelper) repository(path string) string {
	uri := fedoraURI(path)

	repo := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.4.jsonld",
		"@id" : "%s",
		"@type" : "Repository"
	  }
	  `, uri)

	r.putResource(uri, repo)
	return uri
}

func (r *resourceHelper) putResource(uri, body string) {
	request, err := http.NewRequest(http.MethodPut, uri, strings.NewReader(body))
	if err != nil {
		r.t.Fatalf("Building request failed: %s", err)
	}

	request.Header.Set("Content-Type", "application/ld+json")
	request.Header.Set("Prefer", `handling=lenient; received="minimal"`)
	authz(request)

	resp, err := r.c.Do(request)
	if err != nil {
		r.t.Fatalf("PUT request failed: %s", err)
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode > 299 {
		r.t.Fatalf("Could not add resource: %d, body:\n%s", resp.StatusCode, body)
	}
}
