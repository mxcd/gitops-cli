package templating

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

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

func LoadValues() error {
	log.Trace("Loading values files")
	secretFiles, err := util.GetSecretFiles()
	if err != nil {
		return err
	}
	var valuesFiles []string
	for _, secretFile := range secretFiles {
		if strings.HasSuffix(secretFile, "values.gitops.secret.enc.yaml") || strings.HasSuffix(secretFile, "values.gitops.secret.enc.yml") {
			valuesFiles = append(valuesFiles, secretFile)
		}
	}

	valuesFiles = filterValuesFiles(valuesFiles, util.GetCliContext().String("dir"))

	templateValues = TemplateValues{}

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

func filterValuesFiles(valuesFiles []string, dirLimit string) []string {
	// If dirLimit points to a secret file, extract its directory
	if strings.HasSuffix(dirLimit, ".gitops.secret.enc.yaml") || strings.HasSuffix(dirLimit, ".gitops.secret.enc.yml") {
		dirLimit = filepath.Dir(dirLimit)
		log.Debugf("Converted file-based dir limit to directory: %q", dirLimit)
	}
	
	dirLimitNormalized := normalizeDirPath(dirLimit)
	log.Debugf("Filtering values files with dirLimit=%q normalized=%q", dirLimit, dirLimitNormalized)
	if dirLimitNormalized == "" {
		return valuesFiles
	}

	filtered := make([]string, 0, len(valuesFiles))
	for _, valuesFile := range valuesFiles {
		valuesDir := normalizeDirPath(filepath.Dir(valuesFile))
		include := shouldIncludeValuesFile(valuesDir, dirLimitNormalized)
		log.Debugf("Evaluating values file %q dir=%q normalizedDir=%q include=%v", valuesFile, filepath.Dir(valuesFile), valuesDir, include)
		if include {
			filtered = append(filtered, valuesFile)
		} else {
			log.Debugf("Excluding values file %q due to dir limit %q", valuesFile, dirLimitNormalized)
		}
	}
	return filtered
}

func shouldIncludeValuesFile(valuesDir string, dirLimit string) bool {
	if dirLimit == "" {
		return true
	}
	if valuesDir == "" {
		return true
	}
	return strings.HasPrefix(valuesDir, dirLimit) || strings.HasPrefix(dirLimit, valuesDir)
}

func normalizeDirPath(dir string) string {
	dir = filepath.ToSlash(strings.TrimSpace(dir))
	dir = strings.TrimPrefix(dir, "./")
	dir = strings.Trim(dir, "/")
	if dir == "" || dir == "." {
		return ""
	}
	return dir + "/"
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
	order := make([]string, len(t))
	for idx, tv := range t {
		order[idx] = tv.Path
	}
	log.Debugf("TemplateValues merge order: %v", order)
	for i, templateValue := range t {
		var bestParent *TemplateValuesPath
		for j := i - 1; j >= 0; j-- {
			candidate := t[j]
			log.Debugf("Considering ancestor candidate %q for child %q", candidate.Path, templateValue.Path)
			if strings.HasPrefix(templateValue.Path, candidate.Path) {
				bestParent = candidate
				break
			}
		}

		if bestParent != nil {
			log.Debugf("Merging ancestor %q into %q", bestParent.Path, templateValue.Path)
			templateValue.MergedValues = mergeMaps(bestParent.MergedValues, templateValue.Values)
		} else {
			log.Debugf("No ancestor found for %q, using own values only", templateValue.Path)
			templateValue.MergedValues = templateValue.Values
		}
	}
}

func GetValuesForPath(path string) map[interface{}]interface{} {
	log.Debugf("Resolving values for secret path %q", path)
	if !loaded {
		err := LoadValues()
		if err != nil {
			log.Panic(err)
		}
	}
	values := map[interface{}]interface{}{}
	usedPath := ""
	maxPathLength := 0
	for _, templateValue := range templateValues {
		log.Debugf("Considering values prefix %q for path %q", templateValue.Path, path)
		if strings.HasPrefix(path, templateValue.Path) && len(templateValue.Path) > maxPathLength {
			maxPathLength = len(templateValue.Path)
			usedPath = templateValue.Path
			values = templateValue.MergedValues
			log.Debugf("Selecting values prefix %q for path %q", templateValue.Path, path)
		}
	}

	if usedPath != "" {
		log.Debugf("Using values from %s for path %s", usedPath, path)
	} else {
		log.Debugf("No values found for path %s", path)
	}

	return values
}

func TestTemplating(c *cli.Context) error {
	secretFiles, err := util.GetSecretFiles()
	if err != nil {
		log.Fatal(err)
	}
	for _, secretFile := range secretFiles {
		log.Debug(secretFile)
		decryptedFile, err := util.DecryptFile(secretFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Trace(string(decryptedFile))
	}
	return nil
}
