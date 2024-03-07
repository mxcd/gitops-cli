package templating

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mxcd/gitops-cli/internal/util"
	"gopkg.in/yaml.v2"
)

type TemplateValues []*TemplateValuesPath

var templateValues = TemplateValues{}

type TemplateValuesPath struct {
	Path         string
	Values       map[interface{}]interface{}
	MergedValues map[interface{}]interface{}
}

var loaded = false

func LoadValues(directoryLimit string) error {
	log.Trace("Loading values files")
	secretFiles, err := util.GetSecretFiles()
	if err != nil {
		return err
	}
	var valuesFiles []string
	for _, secretFileName := range secretFiles {
		if !strings.HasSuffix(secretFileName, "values.gitops.secret.enc.yaml") && !strings.HasSuffix(secretFileName, "values.gitops.secret.enc.yml") {
			continue
		}
		if !valuesFileApplicable(secretFileName, directoryLimit) {
			log.Trace("Skipping values file due to directory filter: ", secretFileName)
			continue
		}
		valuesFiles = append(valuesFiles, secretFileName)
	}

	for _, valuesFile := range valuesFiles {
		log.Trace("Loading secret values file: ", valuesFile)
		absoluteSecretPath := path.Join(util.GetRootDir(), valuesFile)
		decryptedFileContent, err := util.DecryptFile(absoluteSecretPath)
		if err != nil {
			return err
		}
		var values map[interface{}]interface{}
		yaml.UnmarshalStrict(decryptedFileContent, &values)
		templateValues = append(templateValues, &TemplateValuesPath{
			Path:   fmt.Sprintf("%s/", filepath.ToSlash(filepath.Dir(valuesFile))),
			Values: values,
		})
	}

	templateValues.merge()
	loaded = true
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

func (t TemplateValues) merge() {
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

func GetValuesForPath(path string, directoryLimit string) map[interface{}]interface{} {
	if !loaded {
		err := LoadValues(directoryLimit)
		if err != nil {
			log.Panic(err)
		}
	}
	values := map[interface{}]interface{}{}
	usedPath := ""
	maxPathLength := 0
	for _, templateValue := range templateValues {
		if strings.HasPrefix(path, templateValue.Path) && len(templateValue.Path) > maxPathLength {
			maxPathLength = len(templateValue.Path)
			usedPath = templateValue.Path
			values = templateValue.MergedValues
		}
	}

	if usedPath != "" {
		log.Tracef("Using values from %s for path %s", usedPath, path)
	} else {
		log.Tracef("No values found for path %s", path)
	}

	return values
}

func valuesFileApplicable(secretFile, directoryLimit string) bool {
	secretFile = filepath.Clean(secretFile)
	directoryLimit = filepath.Clean(directoryLimit)

	if directoryLimit == "" {
		return true
	}

	secretFileParentDir := filepath.Dir(secretFile)
	secretFileDirectoryElements := strings.Split(secretFileParentDir, string(filepath.Separator))
	directoryLimitElements := strings.Split(directoryLimit, string(filepath.Separator))

	if len(secretFileDirectoryElements) <= len(directoryLimitElements) {
		for i, secretFileDirectoryElement := range secretFileDirectoryElements {
			if secretFileDirectoryElement != directoryLimitElements[i] {
				return false
			}
		}
	} else {
		for i, directoryLimitElement := range directoryLimitElements {
			if directoryLimitElement != secretFileDirectoryElements[i] {
				return false
			}
		}
	}

	return true
}
