package rule

import (
	"strings"

	"github.com/pkg/errors"
)

// Condition in the policy rules DSL is a json object that determines whether a policy applies or not.
// It is generally of the form
//     {
//        "anyOf": [
//           {"equals":{"one": "two"}},
//           {"endsWith":{"one": "gone"}}
//        ]
//     }
type Condition map[string]interface{}

type evaluation func(interface{}, VariableResolver) (bool, error)

var evaluators map[string]evaluation

func init() {
	evaluators = map[string]evaluation{
		"endsWith": endsWith,
		"equals":   equals,
		"anyOf":    anyOf,
	}
}

// Apply evaluates a condition using the given variable resolver.
func (c Condition) Apply(variables VariableResolver) (bool, error) {

	for cond, val := range c {
		eval, ok := evaluators[cond]
		if !ok {
			return false, errors.Errorf("unknown condition %s", cond)
		}

		passes, err := eval(val, variables)
		if err != nil {
			return false, errors.Wrapf(err, "could not evaluate condition '%s'", cond)
		}

		if !passes {
			return false, nil
		}
	}
	return true, nil
}

func endsWith(fromCondition interface{}, variables VariableResolver) (bool, error) {

	return eachPair(fromCondition, variables, strings.HasSuffix)
}

func equals(fromCondition interface{}, variables VariableResolver) (bool, error) {

	return eachPair(fromCondition, variables, func(a, b string) bool {
		return a == b
	})
}

func anyOf(arg interface{}, variables VariableResolver) (bool, error) {
	list, ok := arg.([]interface{})
	if !ok {
		return false, errors.Errorf("expecting a list, but got %T", arg)
	}

	for _, item := range list {
		obj, ok := item.(map[string]interface{})
		if !ok {
			return false, errors.Errorf("expecting a JSON object as list item, but got %T", arg)
		}

		passes, err := Condition(obj).Apply(variables)
		if err != nil {
			return false, errors.Wrap(err, "condition failed to apply")
		}

		if passes {
			return true, nil
		}
	}

	return false, nil
}

func eachPair(src interface{}, variables VariableResolver, test func(string, string) bool) (passes bool, err error) {
	operands, ok := src.(map[string]interface{})
	if !ok {
		return false, errors.Errorf("expecting a JSON object, instead got a %T", src)
	}

	// If there's no resolver, just pass through.
	if variables == nil {
		variables = passThroughResolver{}
	}

	for b, thing := range operands {
		a, ok := thing.(string)
		if !ok {
			return false, errors.Errorf("given a %T instead of a string", src)
		}

		if a, err = singleValued(variables.Resolve(a)); err != nil {
			return false, errors.Wrapf(err, "could not resolve variable %s", a)
		}
		if b, err = singleValued(variables.Resolve(b)); err != nil {
			return false, errors.Wrapf(err, "could not resolve variable %s", b)
		}

		if !test(a, b) {
			return false, err
		}
	}

	return true, err
}

func singleValued(list []string, err error) (string, error) {
	if err != nil {
		return "", errors.WithStack(err)
	}

	if len(list) == 0 {
		return "", nil
	}

	if len(list) != 1 {
		return "", errors.Errorf("expecting single valued string, instead got %+v", list)
	}

	return list[0], nil
}
