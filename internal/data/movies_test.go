package data

import (
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/validator"
)

func TestValidateMovie(t *testing.T) {
	tests := []struct {
		name  string
		movie Movie
		valid bool
	}{
		{
			name: "valid movie",
			movie: Movie{
				Title:   "Matrix",
				Year:    1999,
				Runtime: 100,
				Genres:  []string{"action"},
			},
			valid: true,
		},
		{
			name: "missing title",
			movie: Movie{
				Year: 1999,
			},
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateMovie(v, &tt.movie)
			if v.Valid() != tt.valid {
				t.Errorf(
					"expected %v",
					tt.valid,
				)
			}
		})
	}
}
