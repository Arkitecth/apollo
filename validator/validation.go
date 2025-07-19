package validator

import (
	"regexp"
	"slices"
)

type Validator struct {
	ErrorMap map[string]string
}

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func New() *Validator {
	return &Validator{
		ErrorMap: make(map[string]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.ErrorMap) == 0
}

func (v *Validator) Exist(key string) bool {
	if _, ok := v.ErrorMap[key]; ok {
		return true
	}
	return false
}

func Matches(value string, regexp *regexp.Regexp) bool {
	return regexp.MatchString(value)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(uniqueValues) == len(values)
}

func (v *Validator) Add(key, value string) {
	if !v.Exist(key) {
		v.ErrorMap[key] = value
	}
}

func (v *Validator) Check(expr bool, key, value string) {
	if expr {
		v.Add(key, value)
	}
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
