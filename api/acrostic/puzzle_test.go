package acrostic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPuzzle_WithoutSolution(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
	}{
		{
			name:   "nil cells",
			puzzle: &Puzzle{Cells: nil},
		},
		{
			name:   "empty cells",
			puzzle: &Puzzle{Cells: [][]string{}},
		},
		{
			name: "non-empty cells",
			puzzle: &Puzzle{
				Cells: [][]string{
					{"A", "B", "C"},
					{"D", "E", "F"},
					{"I", "H", "G"},
				},
			},
		},
		{
			name:   "author",
			puzzle: &Puzzle{Author: "puzzle author"},
		},
		{
			name:   "title",
			puzzle: &Puzzle{Title: "puzzle title"},
		},
		{
			name:   "quote",
			puzzle: &Puzzle{Quote: "puzzle quote"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := test.puzzle.WithoutSolution()

			assert.Nil(t, puzzle.Cells)
			assert.Empty(t, puzzle.Author)
			assert.Empty(t, puzzle.Title)
			assert.Empty(t, puzzle.Quote)
		})
	}
}

func TestPuzzle_GetCellCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		num       int
		expectedX int
		expectedY int
	}{
		{
			name:      "1",
			filename:  "xwordinfo-nyt-20200524.json",
			num:       1,
			expectedX: 0,
			expectedY: 0,
		},
		{
			name:      "2",
			filename:  "xwordinfo-nyt-20200524.json",
			num:       2,
			expectedX: 1,
			expectedY: 0,
		},
		{
			name:      "24",
			filename:  "xwordinfo-nyt-20200524.json",
			num:       24,
			expectedX: 26,
			expectedY: 0,
		},
		{
			name:      "25",
			filename:  "xwordinfo-nyt-20200524.json",
			num:       25,
			expectedX: 0,
			expectedY: 1,
		},
		{
			name:      "100",
			filename:  "xwordinfo-nyt-20200524.json",
			num:       100,
			expectedX: 7,
			expectedY: 4,
		},
		{
			name:      "177",
			filename:  "xwordinfo-nyt-20200524.json",
			num:       177,
			expectedX: 26,
			expectedY: 7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := LoadTestPuzzle(t, test.filename)

			x, y, err := puzzle.GetCellCoordinates(test.num)
			require.NoError(t, err)

			assert.Equal(t, test.expectedX, x)
			assert.Equal(t, test.expectedY, y)
		})
	}
}

func TestPuzzle_GetCellCoordinates_Error(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		num      int
	}{
		{
			name:     "-1",
			filename: "xwordinfo-nyt-20200524.json",
			num:      -1,
		},
		{
			name:     "0",
			filename: "xwordinfo-nyt-20200524.json",
			num:      0,
		},
		{
			name:     "178",
			filename: "xwordinfo-nyt-20200524.json",
			num:      178,
		},
		{
			name:     "10000",
			filename: "xwordinfo-nyt-20200524.json",
			num:      10000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := LoadTestPuzzle(t, test.filename)

			_, _, err := puzzle.GetCellCoordinates(test.num)
			assert.Error(t, err)
		})
	}
}
