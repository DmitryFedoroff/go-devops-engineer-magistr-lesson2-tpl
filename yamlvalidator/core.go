package yamlvalidator

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
)

type Validator struct {
	FilePath       string
	RootNode       yaml.Node
	ErrorCollector ErrorCollector
}

func NewValidator(filePath string) (*Validator, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path: %w", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file content: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("cannot unmarshal file content: %w", err)
	}

	parentDir := filepath.Dir(absPath)
	relPath, _ := filepath.Rel(parentDir, absPath)

	return &Validator{
		FilePath:       relPath,
		RootNode:       root,
		ErrorCollector: ErrorCollector{},
	}, nil
}

func (v *Validator) Validate() []error {
	for _, doc := range v.RootNode.Content {
		v.validatePod(doc)
	}
	return v.ErrorCollector.Errors
}

func (v *Validator) validatePod(node *yaml.Node) {
	requiredFields := []string{"apiVersion", "kind", "metadata", "spec"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		switch key {
		case "apiVersion":
			v.validateAPIVersion(valueNode)
		case "kind":
			v.validateKind(valueNode)
		case "metadata":
			v.validateMetadata(valueNode)
		case "spec":
			v.validateSpec(valueNode)
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validateAPIVersion(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode || node.Value != APIVersionExpected {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("apiVersion has unsupported value '%s'", node.Value),
		})
	}
}

func (v *Validator) validateKind(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode || node.Value != KindExpected {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("kind has unsupported value '%s'", node.Value),
		})
	}
}

func (v *Validator) validateMetadata(node *yaml.Node) {
	requiredFields := []string{"name"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		switch key {
		case "name":
			v.validateName(valueNode, false)
		case "labels":
			v.validateLabels(valueNode)
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validateSpec(node *yaml.Node) {
	requiredFields := []string{"containers"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		switch key {
		case "os":
			v.validateOS(valueNode)
		case "containers":
			v.validateContainers(valueNode)
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validateOS(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "os must be string",
		})
		return
	}

	if !ContainsString(node.Value, SupportedOSNames) {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("os has unsupported value '%s'", node.Value),
		})
	}
}

func (v *Validator) validateContainers(node *yaml.Node) {
	if node.Kind != yaml.SequenceNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "containers must be an array",
		})
		return
	}

	for _, containerNode := range node.Content {
		v.validateContainer(containerNode)
	}
}

func (v *Validator) validateContainer(node *yaml.Node) {
	requiredFields := []string{"name", "image", "resources"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		switch key {
		case "name":
			v.validateName(valueNode, true)
		case "image":
			v.validateImage(valueNode)
		case "ports":
			v.validatePorts(valueNode)
		case "readinessProbe":
			v.validateProbe(valueNode)
		case "livenessProbe":
			v.validateProbe(valueNode)
		case "resources":
			v.validateResources(valueNode)
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validateName(node *yaml.Node, checkSnakeCase bool) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "name must be string",
		})
		return
	}

	if node.Value == "" {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "name is required",
		})
		return
	}

	if checkSnakeCase && !RegexSnakeCase.MatchString(node.Value) {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("name has invalid format '%s'", node.Value),
		})
	}
}

func (v *Validator) validateImage(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "image must be string",
		})
		return
	}

	if !RegexImage.MatchString(node.Value) {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("image has invalid format '%s'", node.Value),
		})
	}
}

func (v *Validator) validatePorts(node *yaml.Node) {
	if node.Kind != yaml.SequenceNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "ports must be an array",
		})
		return
	}

	for _, portNode := range node.Content {
		v.validatePort(portNode)
	}
}

func (v *Validator) validatePort(node *yaml.Node) {
	requiredFields := []string{"containerPort"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		switch key {
		case "containerPort":
			v.validatePortNumber(valueNode, "containerPort")
		case "protocol":
			v.validateProtocol(valueNode)
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validatePortNumber(node *yaml.Node, fieldName string) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("%s must be int", fieldName),
		})
		return
	}

	port, err := strconv.Atoi(node.Value)
	if err != nil {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("%s must be int", fieldName),
		})
		return
	}

	if port < PortNumberMin || port > PortNumberMax {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("%s value out of range", fieldName),
		})
	}
}

func (v *Validator) validateProtocol(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "protocol must be string",
		})
		return
	}

	if !ContainsString(node.Value, SupportedProtocols) {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("protocol has unsupported value '%s'", node.Value),
		})
	}
}

func (v *Validator) validateProbe(node *yaml.Node) {
	requiredFields := []string{"httpGet"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		if key == "httpGet" {
			v.validateHTTPGet(valueNode)
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validateHTTPGet(node *yaml.Node) {
	requiredFields := []string{"path", "port"}
	visitedFields := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value
		visitedFields[key] = true

		switch key {
		case "path":
			v.validateHTTPPath(valueNode)
		case "port":
			v.validatePortNumber(valueNode, "port")
		}
	}

	for _, field := range requiredFields {
		if !visitedFields[field] {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     node.Line,
				Message:  fmt.Sprintf("%s is required", field),
			})
		}
	}
}

func (v *Validator) validateHTTPPath(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "path must be string",
		})
		return
	}

	if !RegexAbsolutePath.MatchString(node.Value) {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("path has invalid format '%s'", node.Value),
		})
	}
}

func (v *Validator) validateResources(node *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		_, valueNode := node.Content[i], node.Content[i+1]
		v.validateResourceRequirements(valueNode)
	}
}

func (v *Validator) validateResourceRequirements(node *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		keyNode, valueNode := node.Content[i], node.Content[i+1]
		key := keyNode.Value

		switch key {
		case "cpu":
			v.validateCPU(valueNode)
		case "memory":
			v.validateMemory(valueNode)
		}
	}
}

func (v *Validator) validateCPU(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "cpu must be int",
		})
		return
	}

	cpu, err := strconv.Atoi(node.Value)
	if err != nil || cpu < 1 {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "cpu value out of range",
		})
	}
}

func (v *Validator) validateMemory(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "memory must be string",
		})
		return
	}

	matches := RegexMemory.FindStringSubmatch(node.Value)
	if matches == nil {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  fmt.Sprintf("memory has invalid format '%s'", node.Value),
		})
		return
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil || amount < 1 {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "memory value out of range",
		})
	}
}

func (v *Validator) validateLabels(node *yaml.Node) {
	if node.Kind != yaml.MappingNode {
		v.ErrorCollector.Add(&ValidationError{
			FilePath: v.FilePath,
			Line:     node.Line,
			Message:  "labels must be a mapping",
		})
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		_, valueNode := node.Content[i], node.Content[i+1]
		if valueNode.Kind != yaml.ScalarNode {
			v.ErrorCollector.Add(&ValidationError{
				FilePath: v.FilePath,
				Line:     valueNode.Line,
				Message:  "label value must be string",
			})
		}
	}
}
