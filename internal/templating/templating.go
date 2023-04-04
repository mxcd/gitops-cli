package templating

import (
	"path"
	"sort"
	"strings"

	"github.com/mxcd/gitops-cli/internal/util"
	"gopkg.in/yaml.v2"
)

type TemplateValues []*TemplateValuesPath
var templateValues = TemplateValues{}

type TemplateValuesPath struct {
	Path          string
	Values        map[interface{}]interface{}
	MergedValues  map[interface{}]interface{}
}

func Load() error {
	secretFiles, err := util.GetSecretFiles(util.GetRootDir())
	if err != nil {
		return err
	}
	var valuesFiles []string
	for _, secretFile := range secretFiles {
		if strings.HasSuffix(secretFile, "values.gitops.secret.enc.yaml") || strings.HasSuffix(secretFile, "values.gitops.secret.enc.yml") {
			valuesFiles = append(valuesFiles, secretFile)
		}
	}

	for _, valuesFile := range valuesFiles {
		absoluteSecretPath := path.Join(util.GetRootDir(), valuesFile)
		decryptedFileContent, err := util.DecryptFile(absoluteSecretPath)
		if err != nil {
			return err
		}
		var values map[interface{}]interface{}
		yaml.UnmarshalStrict(decryptedFileContent, &values)
		templateValues = append(templateValues, &TemplateValuesPath{
			Path:    valuesFile,
			Values:  values,
		})
	}
	return nil
}

func mergeMaps(a, b map[interface{}]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{}, len(a))
	for k, v := range a {
			out[k] = v
	}
	for k, v := range b {
			// If you use map[string]interface{}, ok is always false here.
			// Because yaml.Unmarshal will give you map[interface{}]interface{}.
			if v, ok := v.(map[interface{}]interface{}); ok {
					if bv, ok := out[k]; ok {
							if bv, ok := bv.(map[interface{}]interface{}); ok {
									out[k] = mergeMaps(bv, v)
									continue
							}
					}
			}
			out[k] = v
	}
	return out
}

func (t TemplateValues) Merge() {
	sort.SliceStable(t, func(i, j int) bool {
		return len(strings.Split(t[i].Path, "/")) < len(strings.Split(t[j].Path, "/"))
	})
	for i, templateValue := range t {
		merged := false
		for j, templateValue2 := range t {
			if j >= i {
				break
			}

			if strings.HasPrefix(templateValue.Path, templateValue2.Path) {
				merged = true
				templateValue.MergedValues = mergeMaps(templateValue2.MergedValues, templateValue.Values)
				break
			}
		}
		if !merged {
			templateValue.MergedValues = templateValue.Values
		}
	}
}