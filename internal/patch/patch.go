package patch

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/mxcd/gitops-cli/internal/git"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func Patch(c *cli.Context) error {
	branch := c.String("branch")
	basicAuth := c.String("basicauth")

	var auth transport.AuthMethod = nil

	if basicAuth != "" {
		split := strings.Split(basicAuth, ":")
		if len(split) != 2 {
			return errors.New("invalid basic auth string")
		}
		auth = &http.BasicAuth{
			Username: split[0],
			Password: split[1],
		}
	}

	repository := c.String("repo")
	if repository == "" {
		return errors.New("no repository specified")
	}

	filePath := c.Args().First()
	if filePath == "" {
		return errors.New("no file specified for patching")
	}

	selector := c.Args().Get(1)
	value := c.Args().Get(2)

	repo, err := git.Clone(&git.GitOpsCloneOptions{
		Repository: repository,
		Branch:     branch,
		Auth:       auth,
	})

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Error("Failed to get worktree: ", err)
		return err
	}

	file, err := worktree.Filesystem.Open(filePath)
	if err != nil {
		log.Error("Failed to open file: ", err)
		return err
	}
	defer file.Close()

	// Read the contents of the file
	fileContents, err := io.ReadAll(file)
	if err != nil {
		log.Error("Failed to read file: ", err)
		return err
	}

	log.Debug("Read file ", filePath, ":\n", string(fileContents))

	// TODO patch single file
	var fileData yaml.Node
	err = yaml.Unmarshal(fileContents, &fileData)
	if err != nil {
		panic(err)
	}
	log.Debug("File data: ", fileData)

	patchedData, err := patchYamlData(&fileData, selector, value)
	if err != nil {
		return err
	}
	log.Debug("Patched file data: ", patchedData)

	patchedFileContents, err := yaml.Marshal(patchedData)
	if err != nil {
		return err
	}
	log.Debug("Patched file contents: ", string(patchedFileContents))

	// TODO commit and push

	return err
}

func patchYamlString(yamlString string, selector string, value string) (string, error) {
	var fileData yaml.Node
	err := yaml.Unmarshal([]byte(yamlString), &fileData)
	if err != nil {
		return "", err
	}

	var patchedData *yaml.Node
	if strings.HasPrefix(selector, ".") {
		selector = selector[1:]
		patchedData, err = patchYamlData(&fileData, selector, value)
		if err != nil {
			return "", err
		}
	} else {
		patchedData, err = patchYamlDataWithSearch(&fileData, selector, value)
		if err != nil {
			return "", err
		}
	}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(patchedData)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func patchYamlData(data *yaml.Node, selector string, value string) (*yaml.Node, error) {
	selectorParts := strings.Split(selector, ".")

	if err := findAndPatchNode(data, selectorParts, value); err != nil {
		return nil, err
	}
	return data, nil
}

func findAndPatchNode(node *yaml.Node, selectorParts []string, value string) error {
	if len(selectorParts) == 0 {
		node.Value = value
		return nil
	}

	for i, child := range node.Content {
		if child.Kind == yaml.MappingNode {
			for j := 0; j < len(child.Content); j += 2 {
				keyNode := child.Content[j]
				valueNode := child.Content[j+1]
				if keyNode.Value == selectorParts[0] {
					return findAndPatchNode(valueNode, selectorParts[1:], value)
				}
			}
		} else if child.Kind == yaml.ScalarNode {
			if child.Value == selectorParts[0] {
				return findAndPatchNode(node.Content[i+1], selectorParts[1:], value)
			}
		}
	}

	return errors.New("selector not found")
}

func patchYamlDataWithSearch(data *yaml.Node, selector string, value string) (*yaml.Node, error) {
	if err := findAndPatchNodeWithSearch(data, selector, value); err != nil {
		return nil, err
	}
	return data, nil
}

func findAndPatchNodeWithSearch(node *yaml.Node, selector string, value string) error {
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			if keyNode.Kind == yaml.ScalarNode && keyNode.Value == selector {
				valueNode.Value = value
				return nil
			}
			if err := findAndPatchNodeWithSearch(valueNode, selector, value); err == nil {
				return nil
			}
		}
	} else if node.Kind == yaml.SequenceNode || node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			if err := findAndPatchNodeWithSearch(child, selector, value); err == nil {
				return nil
			}
		}
	}
	return errors.New("key not found")
}
