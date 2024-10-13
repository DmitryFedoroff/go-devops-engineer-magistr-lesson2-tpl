package yamlvalidator

import "regexp"

const (
	APIVersionExpected = "v1"
	KindExpected       = "Pod"
)

var (
	RegexSnakeCase    = regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
	RegexImage        = regexp.MustCompile(`^registry\.bigbrother\.io/(.+):(.+)$`)
	RegexMemory       = regexp.MustCompile(`^(\d+)(Mi|Gi|Ki)$`)
	RegexAbsolutePath = regexp.MustCompile(`^/.*`)

	SupportedOSNames   = []string{"linux", "windows"}
	SupportedProtocols = []string{"TCP", "UDP"}

	PortNumberMin = 1
	PortNumberMax = 65535
)
