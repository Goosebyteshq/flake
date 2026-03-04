package parsers

import (
	"fmt"
	"strings"
)

type Framework string

const (
	FrameworkAuto     Framework = "auto"
	FrameworkGo       Framework = "go"
	FrameworkPytest   Framework = "pytest"
	FrameworkJest     Framework = "jest"
	FrameworkJunitXML Framework = "junitxml"
	FrameworkTAP      Framework = "tap"
	FrameworkCargo    Framework = "cargo"
	FrameworkTRX      Framework = "trx"
	FrameworkSurefire Framework = "surefire"
	FrameworkGradle   Framework = "gradle"
	FrameworkNUnitXML Framework = "nunitxml"
	FrameworkMocha    Framework = "mocha"
)

func ParseFramework(input string) (Framework, error) {
	n := strings.ToLower(strings.TrimSpace(input))
	if n == "" {
		return FrameworkAuto, nil
	}
	if n == string(FrameworkAuto) {
		return FrameworkAuto, nil
	}
	f, ok := registeredAlias(n)
	if !ok {
		return "", fmt.Errorf("unsupported framework %q", input)
	}
	return f, nil
}
