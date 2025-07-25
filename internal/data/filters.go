package data

import (
	"strings"

	"github.com/Arkitecth/apollo/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page < 0, "page", "must be greater than zero")
	v.Check(f.Page >= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize == 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize >= 100, "page_size", "must be a maximum of 100")
	v.Check(!validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
