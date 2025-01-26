package yaml

import (
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

func PatchYaml(yamlData []byte, selector string, value string) ([]byte, error) {
	var node yaml.Node
	err := yaml.Unmarshal(yamlData, &node)
	if err != nil {
		return []byte{}, err
	}

	var patchedNode *yaml.Node
	if strings.HasPrefix(selector, ".") {
		selector = selector[1:]
		patchedNode, err = patchYamlData(&node, selector, value)
		if err != nil {
			return []byte{}, err
		}
	} else {
		patchedNode, err = patchYamlDataWithSearch(&node, selector, value)
		if err != nil {
			return []byte{}, err
		}
	}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(patchedNode)
	if err != nil {
		return []byte{}, err
	}

	return b.Bytes(), nil
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
