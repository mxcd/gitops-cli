package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	originalYaml string
	selector     string
	value        string
	expectedYaml string
}

var testCases = []TestCase{
	{
		originalYaml: `foo: bar
`,
		selector: ".foo",
		value:    "baz",
		expectedYaml: `foo: baz
`,
	},
	{
		originalYaml: `some:
  nested:
    structure:
      foo: bar
`,
		selector: ".some.nested.structure.foo",
		value:    "baz",
		expectedYaml: `some:
  nested:
    structure:
      foo: baz
`,
	},
	{
		originalYaml: `some:
  nested:
    # some comment
    structure:
      foo: bar
`,
		selector: ".some.nested.structure.foo",
		value:    "baz",
		expectedYaml: `some:
  nested:
    # some comment
    structure:
      foo: baz
`,
	},
	{
		originalYaml: `some:
  nested:
    # some comment
    structure:
      foo: bar
`,
		selector: "foo",
		value:    "baz",
		expectedYaml: `some:
  nested:
    # some comment
    structure:
      foo: baz
`,
	},
	{
		originalYaml: `some:
  nested:
    # some comment
    structure:
      - fizz: buzz
      - foo: bar
`,
		selector: "foo",
		value:    "baz",
		expectedYaml: `some:
  nested:
    # some comment
    structure:
      - fizz: buzz
      - foo: baz
`,
	},
	{
		originalYaml: `some:
  nested:
    # some comment
    fizz:
      buzz: bar
    structure:
      foo: bar
`,
		selector: ".some.nested.structure.foo",
		value:    "baz",
		expectedYaml: `some:
  nested:
    # some comment
    fizz:
      buzz: bar
    structure:
      foo: baz
`,
	},
}

func TestYamlPatching(t *testing.T) {
	test := func(testCase TestCase) {

		pathedYamlString, err := PatchYaml([]byte(testCase.originalYaml), testCase.selector, testCase.value)
		assert.NoError(t, err)

		assert.Equal(t, testCase.expectedYaml, string(pathedYamlString), "Patched data should be equal to expected data")
	}

	for _, testCase := range testCases {
		test(testCase)
	}
}
