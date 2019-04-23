//go:generate go run github.com/gobuffalo/packr/v2/packr2

package rule

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"github.com/qri-io/jsonschema"
)

type validationErrors []jsonschema.ValError

func (v validationErrors) Error() string {

	var b strings.Builder
	for _, e := range v {
		b.WriteString(e.Error() + "\n")
	}

	return b.String()
}

// Validate validates a serialized policy rules document with respect to
// the implementation's expected JSON schema, and attempts to parse it.
func Validate(rulesDoc []byte) (*DSL, error) {
	box := packr.New("myBox", "../schemas")
	file, err := box.Open("policy_config_1.0.json")
	if err != nil {
		// Will only happen if internal schema is misnamed
		return nil, errors.Wrapf(err, "could not read schema")
	}

	rs := &jsonschema.RootSchema{}
	err = json.NewDecoder(file).Decode(rs)
	if err != nil {
		// WIll only happen if internal schema is malformed
		return nil, errors.Wrapf(err, "could not parse schema")
	}

	valErrors, err := rs.ValidateBytes(rulesDoc)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse rules doc")
	}

	if len(valErrors) > 0 {
		return nil, validationErrors(valErrors)
	}

	// Serialize just to defend against programmer error
	rules := DSL{}
	decoder := json.NewDecoder(bytes.NewReader(rulesDoc))
	decoder.DisallowUnknownFields()

	return &rules, decoder.Decode(&rules)
}
