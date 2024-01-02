package util

import (
	"fmt"
	"strings"
)

type EnumVar struct {
	Allowed  []string
	variable *string
}

// NewEnumVar give a list of allowed flag parameters
func NewEnumVar(allowed []string, variable *string) *EnumVar {
	if variable == nil {
		panic("variable should not be nil")
	}
	return &EnumVar{
		Allowed:  allowed,
		variable: variable,
	}
}

func (a *EnumVar) Default(d string) *EnumVar {
	*a.variable = d
	return a
}

func (a EnumVar) String() string {
	return *a.variable
}

func (a *EnumVar) Set(p string) error {
	isIncluded := func(opts []string, val string) bool {
		for _, opt := range opts {
			if val == opt {
				return true
			}
		}
		return false
	}
	if !isIncluded(a.Allowed, p) {
		return fmt.Errorf("%s is not included in %s", p, strings.Join(a.Allowed, ","))
	}
	*a.variable = p
	return nil
}

func (a *EnumVar) Type() string {
	return "string"
}
