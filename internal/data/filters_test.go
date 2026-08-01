package data

import (
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/validator"
)

func TestCalculateMetadata(t *testing.T) {
	tests := []struct {
		name         string
		totalRecords int
		page         int
		pageSize     int
		expected     Metadata
	}{
		{
			name:         "empty records",
			totalRecords: 0,
			page:         1,
			pageSize:     20,
			expected:     Metadata{},
		},
		{
			name:         "one page",
			totalRecords: 10,
			page:         1,
			pageSize:     20,
			expected: Metadata{
				CurrentPage:  1,
				PageSize:     20,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 10,
			},
		},
		{
			name:         "multiple pages",
			totalRecords: 45,
			page:         2,
			pageSize:     20,
			expected: Metadata{
				CurrentPage:  2,
				PageSize:     20,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 45,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMetadata(
				tt.totalRecords,
				tt.page,
				tt.pageSize,
			)
			if got != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

func TestFiltersSortColumn(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		safelist []string
		expected string
	}{
		{
			name:     "ascending sort",
			sort:     "title",
			safelist: []string{"title", "year", "-title", "-year"},
			expected: "title",
		},
		{
			name:     "descending sort",
			sort:     "-year",
			safelist: []string{"title", "year", "-title", "-year"},
			expected: "year",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filters{
				Sort:         tt.sort,
				SortSafelist: tt.safelist,
			}
			got := f.sortColumn()
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestFiltersSortDirection(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		expected string
	}{
		{
			name:     "ascending",
			sort:     "title",
			expected: "ASC",
		},
		{
			name:     "descending",
			sort:     "-year",
			expected: "DESC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filters{
				Sort: tt.sort,
			}
			got := f.sortDirection()
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestFiltersLimit(t *testing.T) {
	f := Filters{
		PageSize: 25,
	}
	if got := f.limit(); got != 25 {
		t.Errorf("expected 25, got %d", got)
	}
}

func TestFiltersOffset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{
			page:     1,
			pageSize: 20,
			expected: 0,
		},
		{
			page:     2,
			pageSize: 20,
			expected: 20,
		},
		{
			page:     3,
			pageSize: 10,
			expected: 20,
		},
	}
	for _, tt := range tests {
		f := Filters{
			Page:     tt.page,
			PageSize: tt.pageSize,
		}
		if got := f.offset(); got != tt.expected {
			t.Errorf("expected %d, got %d", tt.expected, got)
		}
	}
}

func TestValidateFilters(t *testing.T) {
	tests := []struct {
		name      string
		filters   Filters
		hasErrors bool
	}{
		{
			name: "valid filters",
			filters: Filters{
				Page:         1,
				PageSize:     20,
				Sort:         "title",
				SortSafelist: []string{"title", "year"},
			},
			hasErrors: false,
		},
		{
			name: "invalid page",
			filters: Filters{
				Page:         0,
				PageSize:     20,
				Sort:         "title",
				SortSafelist: []string{"title"},
			},
			hasErrors: true,
		},
		{
			name: "invalid page size",
			filters: Filters{
				Page:         1,
				PageSize:     200,
				Sort:         "title",
				SortSafelist: []string{"title"},
			},
			hasErrors: true,
		},
		{
			name: "invalid sort",
			filters: Filters{
				Page:         1,
				PageSize:     20,
				Sort:         "invalid",
				SortSafelist: []string{"title"},
			},
			hasErrors: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateFilters(v, tt.filters)
			if v.Valid() != !tt.hasErrors {
				t.Errorf(
					"expected valid=%v, got valid=%v",
					!tt.hasErrors,
					v.Valid(),
				)
			}
		})
	}
}
