package templating

import (
	"testing"

	"github.com/mxcd/gitops-cli/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestMapMerge1(t *testing.T) {
	a := map[interface{}]interface{}{
		"foo": "bar",
		"fizz": "buzz",
	}
	b := map[interface{}]interface{}{
		"fizz": "fizz",
	}
	c := map[interface{}]interface{}{
		"foo": "bar",
		"fizz": "fizz",
	}
	d := mergeMaps(a, b)
	assert.Equal(t, c, d, "Maps should be equal")	
}

func TestMapMerge2(t *testing.T) {
	a := map[interface{}]interface{}{
		"a": "1",
		"b": "2",
	}
	b := map[interface{}]interface{}{
		"c": "3",
		"d": "4",
	}
	c := map[interface{}]interface{}{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
	}
	d := mergeMaps(a, b)
	assert.Equal(t, c, d, "Maps should be equal")	
}

func TestMapMerge3(t *testing.T) {
	a := map[interface{}]interface{}{
		"a": "1",
		"b": []interface{}{"2", "3"},
	}
	b := map[interface{}]interface{}{
		"b": 42,
	}
	c := map[interface{}]interface{}{
		"a": "1",
		"b": 42,
	}
	d := mergeMaps(a, b)
	assert.Equal(t, c, d, "Maps should be equal")	
}

func TestMapMerge4(t *testing.T) {
	a := map[interface{}]interface{}{
		"a": "1",
		"b": 42,
	}
	b := map[interface{}]interface{}{
		"b": []interface{}{"2", "3"},
	}
	c := map[interface{}]interface{}{
		"a": "1",
		"b": []interface{}{"2", "3"},
	}
	d := mergeMaps(a, b)
	assert.Equal(t, c, d, "Maps should be equal")	
}

func TestTemplateValuesMerge(t *testing.T) {
	templateValues := TemplateValues{
		&TemplateValuesPath{
			Path: "fooboo/",
			Values: map[interface{}]interface{}{
				"a": "1",
				"b": "2",
			},
		},
		&TemplateValuesPath{
			Path: "fooboo/bar/",
			Values: map[interface{}]interface{}{
				"a": "11",
				"c": "33",
			},
		},
		&TemplateValuesPath{
			Path: "fizz/",
			Values: map[interface{}]interface{}{
				"a": "5",
				"b": "6",
			},
		},
		&TemplateValuesPath{
			Path: "fizz/buzz/",
			Values: map[interface{}]interface{}{
				"a": "55",
				"c": "77",
			},
		},
	}

	expectedMergedTemplateValues := TemplateValues{
		&TemplateValuesPath{
			Path: "fooboo/",
			Values: map[interface{}]interface{}{
				"a": "1",
				"b": "2",
			},
			MergedValues: map[interface{}]interface{}{
				"a": "1",
				"b": "2",
			},
		},
		&TemplateValuesPath{
			Path: "fizz/",
			Values: map[interface{}]interface{}{
				"a": "5",
				"b": "6",
			},
			MergedValues: map[interface{}]interface{}{
				"a": "5",
				"b": "6",
			},
		},
		&TemplateValuesPath{
			Path: "fooboo/bar/",
			Values: map[interface{}]interface{}{
				"a": "11",
				"c": "33",
			},
			MergedValues: map[interface{}]interface{}{
				"a": "11",
				"b": "2",
				"c": "33",
			},
		},
		&TemplateValuesPath{
			Path: "fizz/buzz/",
			Values: map[interface{}]interface{}{
				"a": "55",
				"c": "77",
			},
			MergedValues: map[interface{}]interface{}{
				"a": "55",
				"b": "6",
				"c": "77",
			},
		},
	}

	templateValues.merge()
	assert.Equal(t, expectedMergedTemplateValues, templateValues, "TemplateValues should be equal")
}


func TestValuesFileLoading(t *testing.T) {
	c := util.GetDummyCliContext()
	util.SetCliContext(c)
	util.ComputeRootDir(c)
	
	LoadValues()
	
	valuesSet1 := map[interface{}]interface{}{
		"namespace": "gitops-dev",
		"stage": "dev",
		"databaseUsername": "my-very-strong-username",
		"databasePassword": "my-very-strong-password",
	}

	mergedValues1 := GetValuesForPath("test_assets/implicit-name.gitops.secret.enc.yml")
	assert.Equal(t, valuesSet1, mergedValues1, "Values should be equal")

	valuesSet2 := map[interface{}]interface{}{
		"namespace": "gitops-dev",
		"stage": "sub-dev",
		"databaseUsername": "my-very-strong-username",
		"databasePassword": "my-very-strong-password",
		"key": "fooo",
	}
	mergedValues2 := GetValuesForPath("test_assets/subdirectory/subdir-secret.gitops.secret.enc.yml")
	assert.Equal(t, valuesSet2, mergedValues2, "Values should be equal")
}