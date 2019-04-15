package main

import (
	"os"
	"testing"
)

func TestValidateCLI(t *testing.T) {
	cases := []struct {
		name        string
		file        string
		errExpected bool
	}{
		{"goodData", "../../rule/testdata/good.json", false},
		{"badData", "../../rule/testdata/bad.json", true},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {

			// Invoke the main function with args manually
			os.Args = []string{"pass-policy-service", "validate", c.file}

			// If it exits in error, capture the error
			var err error
			fatalf = func(f string, a ...interface{}) {
				e, ok := a[len(a)-1].(error)
				if ok {
					err = e
				}
			}

			main()

			if (err != nil) != c.errExpected {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				t.Fatalf("got unexpected error %+v", err)
			}
		})
	}
}
