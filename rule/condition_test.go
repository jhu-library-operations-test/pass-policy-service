package rule_test

import (
	"encoding/json"
	"testing"

	"github.com/oa-pass/pass-policy-service/rule"
)

func TestConditionApply(t *testing.T) {
	variables := testResolver(map[string][]string{
		"${one.spelled}": {"one"},
		"${none}":        {},
	})

	cases := []struct {
		json     string
		expected bool
	}{{
		expected: true,
		json: `{
			"anyOf": [
				{"equals":{"one": "two"}},
				{"endsWith":{"one": "gone"}}
			]
		}`,
	}, {
		expected: false,
		json: `{
			"anyOf": [
				{"equals":{"one": "two"}},
				{"endsWith":{"one": "goner"}}
			]
		}`,
	}, {
		expected: true,
		json: `{
			"noneOf": [
				{"equals":{"one": "two"}},
				{"endsWith":{"one": "goner"}}
			]
		}`,
	}, {
		expected: false,
		json: `{
			"noneOf": [
				{"equals":{"one": "two"}},
				{"endsWith":{"one": "gone"}}
			]
		}`,
	}, {
		expected: false,
		json:     `{"equals":{"one": "two"}}`,
	}, {
		expected: true,
		json:     `{"equals":{"two": "two"}}`,
	}, {
		expected: true,
		json:     `{"contains": {"FACULTY": "STAFF,FACULTY,COW"}}`,
	}, {
		expected: false,
		json:     `{"contains": {"BOVINE": "STAFF,FACULTY,COW"}}`,
	}, {
		expected: true,
		json: `{
			"equals": {"one": "${one.spelled}"},
			"endsWith": {"${one.spelled}": "gone"}
		}`,
	}, {
		expected: false,
		json: `{
			"equals": {"one": "one"},
			"endsWith": {"one": "goner"}
		}`,
	}, {
		expected: false,
		json: `{
				"equals": {"one": "${none}"}
			}`,
	}}

	for _, c := range cases {
		c := c
		t.Run(c.json, func(t *testing.T) {
			parsed := make(map[string]interface{})
			err := json.Unmarshal([]byte(c.json), &parsed)
			if err != nil {
				t.Fatalf("bad test data, does not parse:\n%s\n  reason: %s", c.json, err)
			}

			passed, err := rule.Condition(parsed).Apply(variables)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if passed != c.expected {
				t.Fatalf("passed: %t, but expected %t", passed, c.expected)
			}
		})
	}
}

func TestConditionErrors(t *testing.T) {

	cases := map[string]struct {
		json     string
		resolver rule.VariableResolver
	}{
		"bad JSON": {json: `{
			"good"
			"bad
		}`},
		"no such condition": {json: `{
			"squareRoot": {"2": "4"}
		}`},
		"bad operands": {json: `{
			"equals": {"2": ["3", "5"]}
		}`},
		"bad operands 2": {json: `{
			"equals": 7
		}`},
		"variable resolving error 1": {
			json: `{
				"equals": {"foo": "${bar}"}
			}`,
			resolver: errResolver{},
		},
		"variable resolving error 2": {
			json: `{
				"equals": {"${bar}": "foo"}
			}`,
			resolver: errResolver{},
		},
		"variable resolves to a list": {
			json: `{
				"equals": {"${bar}": "foo"}
			}`,
			resolver: testResolver(map[string][]string{
				"${bar}": {"foo", "bar"},
			}),
		},
		"anyOf is not a list": {
			json: `{
				"anyOf": {"foo": "bar"}
			}`,
		},
		"anyOf is a heterogeneous list": {
			json: `{
				"anyOf": [
					{"equals":{"one": "two"}},
					"hello"
				]
			}`,
		},
		"anyOf resolving error": {
			json: `{
				"anyOf": [
					{"equals":{"one": "two"}},
					{"endsWith":{"one": "${bar}"}}
				]
			}`,
			resolver: errResolver{},
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			parsed := make(map[string]interface{})

			err := json.Unmarshal([]byte(c.json), &parsed)
			if err == nil {
				_, err := rule.Condition(parsed).Apply(c.resolver)
				if err == nil {
					t.Fatal("should have failed with an error")
				}
			}
		})
	}
}
