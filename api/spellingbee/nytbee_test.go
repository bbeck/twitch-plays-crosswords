package spellingbee

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestInferPuzzle(t *testing.T) {
	tests := []struct {
		name       string
		official   []string
		unofficial []string
		center     string
		letters    []string
	}{
		{
			name: "nytbee-20200408 words",
			official: []string{
				"COCONUT",
				"CONCOCT",
				"CONTORT",
				"CONTOUR",
				"COOT",
				"COTTON",
				"COTTONY",
				"COUNT",
				"COUNTRY",
				"COUNTY",
				"COURT",
				"CROUTON",
				"CURT",
				"CUTOUT",
				"NUTTY",
				"ONTO",
				"OUTCRY",
				"OUTRO",
				"OUTRUN",
				"ROOT",
				"ROTO",
				"ROTOR",
				"ROUT",
				"RUNOUT",
				"RUNT",
				"RUNTY",
				"RUTTY",
				"TONY",
				"TOON",
				"TOOT",
				"TORN",
				"TORO",
				"TORT",
				"TOUR",
				"TOUT",
				"TROT",
				"TROUT",
				"TROY",
				"TRYOUT",
				"TURN",
				"TURNOUT",
				"TUTOR",
				"TUTU",
				"TYCOON",
				"TYRO",
				"UNCUT",
				"UNTO",
				"YURT",
			},
			unofficial: []string{
				"CONCOCTOR",
				"CONTO",
				"CORNUTO",
				"CROTON",
				"CRYOTRON",
				"CUNT",
				"CUTTY",
				"CYTON",
				"NOCTURN",
				"NONCOUNT",
				"NONCOUNTRY",
				"NONCOUNTY",
				"NOTTURNO",
				"OCTOROON",
				"OTTO",
				"OUTCOUNT",
				"OUTROOT",
				"OUTTROT",
				"OUTTURN",
				"ROOTY",
				"RYOT",
				"TOCO",
				"TORC",
				"TOROT",
				"TORR",
				"TORY",
				"TOTTY",
				"TOUTON",
				"TOYO",
				"TOYON",
				"TROU",
				"TROUTY",
				"TUNNY",
				"TURNON",
				"TURR",
				"TUTTY",
				"UNROOT",
				"UNTORN",
			},
			center:  "T",
			letters: []string{"C", "N", "O", "R", "U", "Y"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle, err := InferPuzzle(test.official, test.unofficial)
			require.NoError(t, err)

			assert.ElementsMatch(t, test.official, puzzle.OfficialAnswers)
			assert.ElementsMatch(t, test.unofficial, puzzle.UnofficialAnswers)
			assert.Equal(t, test.center, puzzle.CenterLetter)
			assert.ElementsMatch(t, test.letters, puzzle.Letters)
		})
	}
}

func TestInferPuzzle_Error(t *testing.T) {
	official := []string{
		"COCONUT",
		"CONCOCT",
		"CONTORT",
		"CONTOUR",
		"COOT",
		"COTTON",
		"COTTONY",
		"COUNT",
		"COUNTRY",
		"COUNTY",
		"COURT",
		"CROUTON",
		"CURT",
		"CUTOUT",
		"NUTTY",
		"ONTO",
		"OUTCRY",
		"OUTRO",
		"OUTRUN",
		"ROOT",
		"ROTO",
		"ROTOR",
		"ROUT",
		"RUNOUT",
		"RUNT",
		"RUNTY",
		"RUTTY",
		"TONY",
		"TOON",
		"TOOT",
		"TORN",
		"TORO",
		"TORT",
		"TOUR",
		"TOUT",
		"TROT",
		"TROUT",
		"TROY",
		"TRYOUT",
		"TURN",
		"TURNOUT",
		"TUTOR",
		"TUTU",
		"TYCOON",
		"TYRO",
		"UNCUT",
		"UNTO",
		"YURT",
	}

	unofficial := []string{
		"CONCOCTOR",
		"CONTO",
		"CORNUTO",
		"CROTON",
		"CRYOTRON",
		"CUNT",
		"CUTTY",
		"CYTON",
		"NOCTURN",
		"NONCOUNT",
		"NONCOUNTRY",
		"NONCOUNTY",
		"NOTTURNO",
		"OCTOROON",
		"OTTO",
		"OUTCOUNT",
		"OUTROOT",
		"OUTTROT",
		"OUTTURN",
		"ROOTY",
		"RYOT",
		"TOCO",
		"TORC",
		"TOROT",
		"TORR",
		"TORY",
		"TOTTY",
		"TOUTON",
		"TOYO",
		"TOYON",
		"TROU",
		"TROUTY",
		"TUNNY",
		"TURNON",
		"TURR",
		"TUTTY",
		"UNROOT",
		"UNTORN",
	}

	tests := []struct {
		name       string
		official   []string
		unofficial []string
	}{
		{
			name:       "no official words",
			unofficial: unofficial,
		},
		{
			name:     "no unofficial words",
			official: official,
		},
		{
			name:       "official word too short",
			official:   append(official, "RUT"),
			unofficial: unofficial,
		},
		{
			name:       "unofficial word too short",
			official:   official,
			unofficial: append(unofficial, "RUT"),
		},
		{
			name:       "multiple options for center letter",
			official:   []string{"COCONUT"},
			unofficial: []string{"COCONUT"},
		},
		{
			name: "no options for center letter",
			official: []string{
				"ABCDE",
				"FGHIJ",
			},
			unofficial: []string{
				"ABCDE",
				"FGHIJ",
			},
		},
		{
			name: "too many possible letters",
			official: []string{
				"ABCD",
				"AFGH",
				"AIJK",
			},
			unofficial: []string{
				"ABCD",
				"AFGH",
				"AIJK",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := InferPuzzle(test.official, test.unofficial)
			assert.Error(t, err)
		})
	}
}

func TestParseNYTBeeResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:  "official answers",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := []string{
					"COCONUT",
					"CONCOCT",
					"CONTORT",
					"CONTOUR",
					"COOT",
					"COTTON",
					"COTTONY",
					"COUNT",
					"COUNTRY",
					"COUNTY",
					"COURT",
					"CROUTON",
					"CURT",
					"CUTOUT",
					"NUTTY",
					"ONTO",
					"OUTCRY",
					"OUTRO",
					"OUTRUN",
					"ROOT",
					"ROTO",
					"ROTOR",
					"ROUT",
					"RUNOUT",
					"RUNT",
					"RUNTY",
					"RUTTY",
					"TONY",
					"TOON",
					"TOOT",
					"TORN",
					"TORO",
					"TORT",
					"TOUR",
					"TOUT",
					"TROT",
					"TROUT",
					"TROY",
					"TRYOUT",
					"TURN",
					"TURNOUT",
					"TUTOR",
					"TUTU",
					"TYCOON",
					"TYRO",
					"UNCUT",
					"UNTO",
					"YURT",
				}
				assert.ElementsMatch(t, expected, puzzle.OfficialAnswers)
			},
		},
		{
			name:  "unofficial answers",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := []string{
					"CONCOCTOR",
					"CONTO",
					"CORNUTO",
					"CROTON",
					"CRYOTRON",
					"CUNT",
					"CUTTY",
					"CYTON",
					"NOCTURN",
					"NONCOUNT",
					"NONCOUNTRY",
					"NONCOUNTY",
					"NOTTURNO",
					"OCTOROON",
					"OTTO",
					"OUTCOUNT",
					"OUTROOT",
					"OUTTROT",
					"OUTTURN",
					"ROOTY",
					"RYOT",
					"TOCO",
					"TORC",
					"TOROT",
					"TORR",
					"TORY",
					"TOTTY",
					"TOUTON",
					"TOYO",
					"TOYON",
					"TROU",
					"TROUTY",
					"TUNNY",
					"TURNON",
					"TURR",
					"TUTTY",
					"UNROOT",
					"UNTORN",
				}
				assert.ElementsMatch(t, expected, puzzle.UnofficialAnswers)
			},
		},
		{
			name:  "letters",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "T", puzzle.CenterLetter)

				expected := []string{"C", "N", "O", "R", "U", "Y"}
				assert.ElementsMatch(t, expected, puzzle.Letters)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle, err := ParseNYTBeeResponse(test.input)
			require.NoError(t, err)
			test.verify(t, puzzle)
		})
	}
}

func TestParseNYTBeeResponse_Error(t *testing.T) {
	tests := []struct {
		name  string
		input io.Reader
	}{
		{
			name:  "reader returning error",
			input: iotest.TimeoutReader(strings.NewReader("random input")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseNYTBeeResponse(test.input)
			require.Error(t, err)
		})
	}
}
