package validator

import "slices"

type Validator struct {
	ErrorMap map[string]string
}

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
