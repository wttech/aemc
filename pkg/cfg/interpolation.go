package cfg

import (
	"fmt"
	"github.com/compose-spec/compose-go/v2/template"
	"gopkg.in/yaml.v3"
	"strings"
)

// interpolateYAML parses a YAML (or JSON) document and replaces variables in its values,
// following the Compose Specification interpolation rules (${VAR}, ${VAR:-default}, ${VAR:?error}, $$ to escape).
//
// Values are interpolated after parsing, as Docker Compose does, so a variable never changes the document structure
// and comments are left alone. An unquoted value is typed after interpolation, so 'active: ${ACTIVE:-true}' is a boolean,
// while a quoted one stays a string.
func interpolateYAML(content []byte, mapping template.Mapping) (map[string]any, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, err
	}
	result := map[string]any{}
	if len(doc.Content) == 0 {
		return result, nil
	}
	if err := interpolateNode(&doc, mapping); err != nil {
		return nil, err
	}
	if err := doc.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func interpolateNode(node *yaml.Node, mapping template.Mapping) error {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 1; i < len(node.Content); i += 2 {
			if err := interpolateNode(node.Content[i], mapping); err != nil {
				return err
			}
		}
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			if err := interpolateNode(child, mapping); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		return interpolateScalar(node, mapping)
	}
	return nil
}

func interpolateScalar(node *yaml.Node, mapping template.Mapping) error {
	if !strings.Contains(node.Value, "$") {
		return nil
	}
	value, err := template.Substitute(node.Value, mapping)
	if err != nil {
		return fmt.Errorf("line %d: %w", node.Line, err)
	}
	node.Value = value
	if node.Style == 0 {
		node.Tag = ""
	}
	return nil
}
