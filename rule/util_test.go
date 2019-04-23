package rule_test

import (
	"encoding/json"
	"fmt"

	"github.com/oa-pass/pass-policy-service/rule"
)

type testResolver map[string][]string

func (t testResolver) Resolve(varString string) ([]string, error) {
	resolved, ok := t[varString]
	if !ok {
		return []string{varString}, nil
	}
	return resolved, nil
}

type errResolver struct{}

func (errResolver) Resolve(varString string) ([]string, error) {
	if rule.IsVariable(varString) {
		return nil, fmt.Errorf("This always returns an error")
	}
	return []string{varString}, nil
}

// map of urls to json strings
type testFetcher map[string]string

// deserialize the json string into the given entity pointer
func (f testFetcher) FetchEntity(url string, entityPointer interface{}) error {
	jsonBlob, ok := f[url]
	if !ok {
		return fmt.Errorf("no value for key %s", url)
	}

	return json.Unmarshal([]byte(jsonBlob), entityPointer)
}

type errFetcher struct{}

func (errFetcher) FetchEntity(url string, entityPointer interface{}) error {
	return fmt.Errorf("This always returns an error")
}
