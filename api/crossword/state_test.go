package crossword

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestState_ApplyAnswer_Cells(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		setup  map[string]string // initial answers applied before the desired answer
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
			},
		},
		{
			name:   "down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
			},
		},
		{
			name:   "overwriting across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "XXXXX",
			},
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
			},
		},
		{
			name:   "overwriting down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1d": "XXXX",
			},
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
			},
		},
		{
			name:   "unknown letters",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: ". AND .",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
			},
		},
		{
			name:   "delete part of answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "Q AND A",
			},
			clue:   "1a",
			answer: ".AND.",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
			},
		},
		{
			name:   "rebus",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "(RED)AND(BLUE)",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "RED", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "BLUE", state.Cells[0][4])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			for clue, answer := range test.setup {
				require.NoError(t, s.ApplyAnswer(clue, answer, false))
			}

			err := s.ApplyAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Cells_CorrectOnly(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		setup  map[string]string // initial answers applied before the desired answer
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
			},
		},
		{
			name:   "down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
			},
		},
		{
			name:   "unknown letters",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: ". AND .",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
			},
		},
		{
			name:   "can fill in some unknown letters",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": ". AND .",
			},
			clue:   "1a",
			answer: "Q AND .",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
			},
		},
		{
			name:   "can fill in all unknown letters",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": ". AND .",
			},
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
			},
		},
		{
			name:   "rebus",
			puzzle: LoadTestPuzzle(t, "xwordinfo-rebus-20181227.json"),
			clue:   "30a",
			answer: "AERIAL RE(CON)",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "A", state.Cells[6][0])
				assert.Equal(t, "E", state.Cells[6][1])
				assert.Equal(t, "R", state.Cells[6][2])
				assert.Equal(t, "I", state.Cells[6][3])
				assert.Equal(t, "A", state.Cells[6][4])
				assert.Equal(t, "L", state.Cells[6][5])
				assert.Equal(t, "R", state.Cells[6][6])
				assert.Equal(t, "E", state.Cells[6][7])
				assert.Equal(t, "CON", state.Cells[6][8])

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			for clue, answer := range test.setup {
				require.NoError(t, s.ApplyAnswer(clue, answer, false))
			}

			err := s.ApplyAnswer(test.clue, test.answer, true)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Cells_CorrectOnly_Error(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		setup  map[string]string // initial answers applied before the desired answer
		clue   string
		answer string
	}{
		{
			name:   "cannot specify incorrect cell",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "R AND A",
		},
		{
			name:   "cannot change correct cell",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "Q AND A",
			},
			clue:   "1a",
			answer: "R AND A",
		},
		{
			name:   "cannot clear correct cell",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "Q AND A",
			},
			clue:   "1a",
			answer: ". AND A",
		},
		{
			name:   "cannot incorrectly specify missing cell",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": ". AND A",
			},
			clue:   "1a",
			answer: "R AND A",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			for clue, answer := range test.setup {
				require.NoError(t, s.ApplyAnswer(clue, answer, false))
			}

			err := s.ApplyAnswer(test.clue, test.answer, true)
			assert.Error(t, err)
		})
	}
}

func TestState_ApplyAnswer_Filled(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		setup  map[string]string // initial answers applied before the desired answer
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.True(t, state.AcrossCluesFilled[1])

				// Everything else should be unfilled
				for num, filled := range state.AcrossCluesFilled {
					if num != 1 {
						assert.False(t, filled)
					}
				}
				for _, filled := range state.DownCluesFilled {
					assert.False(t, filled)
				}
			},
		},
		{
			name:   "down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.True(t, state.DownCluesFilled[1])

				// Everything else should be unfilled
				for _, filled := range state.AcrossCluesFilled {
					assert.False(t, filled)
				}
				for num, filled := range state.DownCluesFilled {
					if num != 1 {
						assert.False(t, filled)
					}
				}
			},
		},
		{
			name:   "across answer completes multiple down clues",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"14a": "THIRD",
				"17a": "IM TOO OLD FOR THIS",
				"19a": "PERU",
				"22a": "DOG TAG",
			},
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.True(t, state.DownCluesFilled[1])
				assert.True(t, state.DownCluesFilled[2])
				assert.True(t, state.DownCluesFilled[3])
				assert.False(t, state.DownCluesFilled[4])
				assert.True(t, state.DownCluesFilled[5])
				assert.False(t, state.DownCluesFilled[6])
				assert.False(t, state.DownCluesFilled[7])
				assert.False(t, state.DownCluesFilled[8])
				assert.False(t, state.DownCluesFilled[9])
				assert.False(t, state.DownCluesFilled[10])
				assert.False(t, state.DownCluesFilled[11])
				assert.False(t, state.DownCluesFilled[12])
				assert.False(t, state.DownCluesFilled[13])
			},
		},
		{
			name:   "down answer completes multiple across clues",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"2d": "AHMED",
				"3d": "NITRO",
				"4d": "DROUGHT",
				"5d": "ADO",
			},
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.True(t, state.AcrossCluesFilled[1])
				assert.True(t, state.AcrossCluesFilled[14])
				assert.False(t, state.AcrossCluesFilled[17])
				assert.True(t, state.AcrossCluesFilled[19])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			for clue, answer := range test.setup {
				require.NoError(t, s.ApplyAnswer(clue, answer, false))
			}

			err := s.ApplyAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Status(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		setup  map[string]string // initial answers applied before the desired answer
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "single answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, StatusSolving, state.Status)
			},
		},
		{
			name:   "complete and correct puzzle",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"6a":  "ATTIC",
				"11a": "HON",
				"14a": "THIRD",
				"15a": "LAID ASIDE",
				"17a": "IM TOO OLD FOR THIS",
				"19a": "PERU",
				"20a": "LEAF",
				"21a": "PEONS",
				"22a": "DOG TAG",
				"24a": "LOL",
				"25a": "HAVE NO OOMPH",
				"30a": "MATTE",
				"33a": "IMPLORED",
				"35a": "ERR",
				"36a": "RANGE",
				"38a": "EMO",
				"39a": "WAIT HERE",
				"42a": "EGYPT",
				"44a": "BOO OFF STAGE",
				"47a": "ERS",
				"48a": "EUGENE",
				"51a": "SHARI",
				"54a": "SINN",
				"56a": "WING",
				"58a": "ITS A ZOO OUT THERE",
				"61a": "STEGOSAUR",
				"62a": "HIT ON",
				"63a": "IPA",
				"64a": "NURSE",
				"65a": "OZONE",
			},
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, StatusComplete, state.Status)
			},
		},
		{
			name:   "complete and incorrect puzzle",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"6a":  "XXXXX",
				"11a": "XXX",
				"14a": "XXXXX",
				"15a": "XXXXXXXXX",
				"17a": "XXXXXXXXXXXXXXX",
				"19a": "XXXX",
				"20a": "XXXX",
				"21a": "XXXXX",
				"22a": "XXXXXX",
				"24a": "XXX",
				"25a": "XXXXXXXXXXX",
				"30a": "XXXXX",
				"33a": "XXXXXXXX",
				"35a": "XXX",
				"36a": "XXXXX",
				"38a": "XXX",
				"39a": "XXXXXXXX",
				"42a": "XXXXX",
				"44a": "XXXXXXXXXXX",
				"47a": "XXX",
				"48a": "XXXXXX",
				"51a": "XXXXX",
				"54a": "XXXX",
				"56a": "XXXX",
				"58a": "XXXXXXXXXXXXXXX",
				"61a": "XXXXXXXXX",
				"62a": "XXXXX",
				"63a": "XXX",
				"64a": "XXXXX",
				"65a": "XXXXX",
			},
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, StatusSolving, state.Status)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			for clue, answer := range test.setup {
				require.NoError(t, s.ApplyAnswer(clue, answer, false))
			}

			err := s.ApplyAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Error(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		clue   string
		answer string
	}{
		{
			name:   "bad clue",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "xyz",
			answer: "ABC",
		},
		{
			name:   "invalid clue",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "199a",
			answer: "ABC",
		},
		{
			name:   "bad answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: ")Q AND A",
		},
		{
			name:   "answer too short",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "ABC",
		},
		{
			name:   "answer too long",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			name:   "answer too short (rebus)",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			clue:   "1a",
			answer: "(Q AND A)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, false)
			assert.Error(t, err)
		})
	}
}

func TestState_ClearIncorrectCells(t *testing.T) {
	tests := []struct {
		name   string
		puzzle *Puzzle
		setup  map[string]string
		verify func(*testing.T, *State)
	}{
		{
			name:   "correct across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "QANDA",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
				assert.True(t, state.AcrossCluesFilled[1])
			},
		},
		{
			name:   "correct down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1d": "QTIP",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
				assert.True(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "partially incorrect across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "QNORA",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
				assert.False(t, state.AcrossCluesFilled[1])
			},
		},
		{
			name:   "partially incorrect down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1d": "QTOP",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
				assert.False(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "completely incorrect across answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "XXXXX",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
				assert.False(t, state.AcrossCluesFilled[1])
			},
		},
		{
			name:   "completely incorrect down answer",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1d": "XXXX",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[1][0])
				assert.Equal(t, "", state.Cells[2][0])
				assert.Equal(t, "", state.Cells[3][0])
				assert.False(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "incorrect across answer clears completed down clue",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "XXXXX",
				"1d": "XTIP",
			},
			verify: func(t *testing.T, state *State) {
				assert.False(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "incorrect down answer clears completed across clue",
			puzzle: LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			setup: map[string]string{
				"1a": "XANDA",
				"1d": "XXXX",
			},
			verify: func(t *testing.T, state *State) {
				assert.False(t, state.AcrossCluesFilled[1])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(test.puzzle)
			for clue, answer := range test.setup {
				require.NoError(t, s.ApplyAnswer(clue, answer, false))
			}

			err := s.ClearIncorrectCells()
			assert.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestParseClue(t *testing.T) {
	tests := []struct {
		clue        string
		expectedNum int
		expectedDir string
	}{
		{
			clue:        "1a",
			expectedNum: 1,
			expectedDir: "a",
		},
		{
			clue:        "10a",
			expectedNum: 10,
			expectedDir: "a",
		},
		{
			clue:        "100a",
			expectedNum: 100,
			expectedDir: "a",
		},
		{
			clue:        "1d",
			expectedNum: 1,
			expectedDir: "d",
		},
		{
			clue:        "10d",
			expectedNum: 10,
			expectedDir: "d",
		},
		{
			clue:        "100d",
			expectedNum: 100,
			expectedDir: "d",
		},
	}

	for _, test := range tests {
		t.Run(test.clue, func(t *testing.T) {
			num, dir, err := ParseClue(test.clue)
			require.NoError(t, err)
			assert.Equal(t, test.expectedNum, num)
			assert.Equal(t, test.expectedDir, dir)
		})
	}
}

func TestParseClue_Error(t *testing.T) {
	tests := []string{
		"",
		"1x",
		"1ad",
		"1a2",
		"1",
		"a",
	}

	for _, clue := range tests {
		t.Run(clue, func(t *testing.T) {
			_, _, err := ParseClue(clue)
			assert.Error(t, err)
		})
	}
}

func TestParseAnswer(t *testing.T) {
	tests := []struct {
		answer   string
		expected []string
	}{
		{
			answer:   "ABCDE",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "abcde",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   " ABCDE",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "ABCDE ",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "ABC DE",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "....S",
			expected: []string{"", "", "", "", "S"},
		},
		{
			answer:   "(RED) VELVET CAKE",
			expected: []string{"RED", "V", "E", "L", "V", "E", "T", "C", "A", "K", "E"},
		},
	}

	for _, test := range tests {
		t.Run(test.answer, func(t *testing.T) {
			actual, err := ParseAnswer(test.answer)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParseAnswer_Error(t *testing.T) {
	tests := []string{
		"",
		"((red) velvet cake)",
		"(red velvet cake",
		"((red) velvet cake",
		"red velvet cake)",
		")red velvet cake",
	}

	for _, answer := range tests {
		t.Run(answer, func(t *testing.T) {
			_, err := ParseAnswer(answer)
			assert.Error(t, err)
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		state    Status
		expected string
	}{
		{
			name:     "created",
			state:    StatusCreated,
			expected: "created",
		},
		{
			name:     "paused",
			state:    StatusPaused,
			expected: "paused",
		},
		{
			name:     "solving",
			state:    StatusSolving,
			expected: "solving",
		},
		{
			name:     "complete",
			state:    StatusComplete,
			expected: "complete",
		},
		{
			name:     "invalid",
			state:    Status(17),
			expected: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.state.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		state    Status
		expected []byte
	}{
		{
			name:     "created",
			state:    StatusCreated,
			expected: []byte(`"created"`),
		},
		{
			name:     "paused",
			state:    StatusPaused,
			expected: []byte(`"paused"`),
		},
		{
			name:     "solving",
			state:    StatusSolving,
			expected: []byte(`"solving"`),
		},
		{
			name:     "complete",
			state:    StatusComplete,
			expected: []byte(`"complete"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs, err := json.Marshal(test.state)
			require.NoError(t, err)
			assert.Equal(t, test.expected, bs)
		})
	}
}

func TestStatus_MarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name  string
		state Status
	}{
		{
			name:  "invalid",
			state: Status(17),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := json.Marshal(test.state)
			assert.Error(t, err)
		})
	}
}

func TestStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected Status
	}{
		{
			name:     "created",
			bs:       []byte(`"created"`),
			expected: StatusCreated,
		},
		{
			name:     "paused",
			bs:       []byte(`"paused"`),
			expected: StatusPaused,
		},
		{
			name:     "solving",
			bs:       []byte(`"solving"`),
			expected: StatusSolving,
		},
		{
			name:     "complete",
			bs:       []byte(`"complete"`),
			expected: StatusComplete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Status

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatus_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "wrong type",
			bs:   []byte(`true`),
		},
		{
			name: "empty value",
			bs:   []byte(`""`),
		},
		{
			name: "invalid value",
			bs:   []byte(`"asdf"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Status

			err := json.Unmarshal(test.bs, &actual)
			assert.Error(t, err)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected []byte
	}{
		{
			name:     "empty",
			duration: Duration{},
			expected: []byte(`"0s"`),
		},
		{
			name:     "1 second",
			duration: Duration{time.Second},
			expected: []byte(`"1s"`),
		},
		{
			name:     "1 minute",
			duration: Duration{time.Minute},
			expected: []byte(`"1m0s"`),
		},
		{
			name:     "1 hour",
			duration: Duration{time.Hour},
			expected: []byte(`"1h0m0s"`),
		},
		{
			name:     "24 hours",
			duration: Duration{24 * time.Hour},
			expected: []byte(`"24h0m0s"`),
		},
		{
			name:     "2 hours 12 minutes 9 seconds",
			duration: Duration{2*time.Hour + 12*time.Minute + 9*time.Second},
			expected: []byte(`"2h12m9s"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := json.Marshal(test.duration)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected Duration
	}{
		{
			name:     "empty",
			bs:       []byte(`"0s"`),
			expected: Duration{},
		},
		{
			name:     "1 second",
			bs:       []byte(`"1s"`),
			expected: Duration{time.Second},
		},
		{
			name:     "1 minute",
			bs:       []byte(`"1m0s"`),
			expected: Duration{time.Minute},
		},
		{
			name:     "1 hour",
			bs:       []byte(`"1h0m0s"`),
			expected: Duration{time.Hour},
		},
		{
			name:     "24 hours",
			bs:       []byte(`"24h0m0s"`),
			expected: Duration{24 * time.Hour},
		},
		{
			name:     "2 hours 12 minutes 9 seconds",
			bs:       []byte(`"2h12m9s"`),
			expected: Duration{2*time.Hour + 12*time.Minute + 9*time.Second},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Duration

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDuration_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "invalid type",
			bs:   []byte(`true`),
		},
		{
			name: "empty value",
			bs:   []byte(`""`),
		},
		{
			name: "incorrect value",
			bs:   []byte(`"1x2y"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Duration

			err := json.Unmarshal(test.bs, &actual)
			assert.Error(t, err)
		})
	}
}

func newState(puzzle *Puzzle) *State {
	var s State
	s.Puzzle = puzzle

	for y := 0; y < s.Puzzle.Rows; y++ {
		s.Cells = append(s.Cells, make([]string, s.Puzzle.Cols))
	}

	s.AcrossCluesFilled = make(map[int]bool)
	s.DownCluesFilled = make(map[int]bool)
	s.Status = StatusSolving

	return &s
}
