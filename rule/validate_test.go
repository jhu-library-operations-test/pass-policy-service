package rule_test

import (
	"io/ioutil"
	"testing"

	"github.com/oa-pass/pass-policy-service/rule"
)

// A known-good document should validate just fine
func TestValidateGoodData(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/good.json")

	_, err := rule.Validate(content)
	if err != nil {
		t.Fatalf("Validation failed: %+v", err)
	}
}

// Known-bad documents should fail
func TestValidateBadData(t *testing.T) {
	invalidDoc, _ := ioutil.ReadFile("testdata/bad.json")

	cases := map[string][]byte{
		"schemaInvalid": invalidDoc,
		"badJSON":       []byte(`{moo`),
	}

	for name, content := range cases {
		content := content
		t.Run(name, func(t *testing.T) {
			_, err := rule.Validate(content)
			if err == nil {
				t.Fatalf("Validation should have failed!")
			}
			_ = err.Error()
		})
	}

}
