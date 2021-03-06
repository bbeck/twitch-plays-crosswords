package spellingbee

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

var Channel = ChannelClient{name: "channel"}

func TestRoute_UpdatePuzzle_NewYorkTimes(t *testing.T) {
	// This acts as a small integration test updating the date of the NYTBee
	// puzzle we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "nytbee-20200408.html")

	response := Channel.PUT("/", `{"new_york_times_date": "2020-04-08"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_JSONError(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int
	}{
		{
			name:     "bad json",
			json:     `{"new_york_times_date": }`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			json:     `{}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)
			ForcePuzzleToBeLoaded(t, "nytbee-20200408.html")

			response := Channel.PUT("/", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdatePuzzle_LoadSaveError(t *testing.T) {
	tests := []struct {
		name                  string
		forcedPuzzleLoadError error
		forcedStateSaveError  error
		expected              int
	}{
		{
			name:                  "nytbee error loading puzzle",
			forcedPuzzleLoadError: errors.New("forced error"),
			expected:              http.StatusInternalServerError,
		},
		{
			name:                  "error saving state",
			forcedPuzzleLoadError: nil,
			forcedStateSaveError:  errors.New("forced error"),
			expected:              http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			if test.forcedPuzzleLoadError != nil {
				ForceErrorDuringPuzzleLoad(t, test.forcedPuzzleLoadError)
			} else {
				ForcePuzzleToBeLoaded(t, "nytbee-20200408.html")
			}

			ForceErrorDuringStateSave(t, test.forcedStateSaveError)

			response := Channel.PUT("/", `{"new_york_times_date": "ignored"}`, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdateSetting(t *testing.T) {
	// This acts as a small integration test updating each setting in turn and
	// making sure the proper value is written to the database and that clients
	// receive events notifying them of the setting change.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Update each setting, one at a time.
	response := Channel.PUT("/setting/allow_unofficial_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.True(t, s.AllowUnofficialAnswers)
	})

	response = Channel.PUT("/setting/font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.Equal(t, model.FontSizeXLarge, s.FontSize)
	})

	response = Channel.PUT("/setting/show_answer_placeholders", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.True(t, s.ShowAnswerPlaceholders)
	})
}

func TestRoute_UpdateSetting_AllowUnofficialAnswers_ClearsAnswers(t *testing.T) {
	// This acts as a small integration test toggling the AllowUnofficialAnswers
	// setting and ensuring that it removes any unofficial answers.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Setup the state with some unofficial answers in it.
	state := NewState(t, "nytbee-20200408.html")
	state.Status = model.StatusSolving
	state.Words = map[string]int{
		"COCONUT":   0,
		"CONCOCT":   1,
		"CONCOCTOR": 2,
		"CONTO":     3,
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Set the AllowUnofficialAnswers setting to false
	response := Channel.PUT("/setting/allow_unofficial_answers", `false`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		expected := map[string]int{
			"COCONUT": 0,
			"CONCOCT": 1,
		}
		assert.Equal(t, expected, state.Words)
	})
}

func TestRoute_UpdateSetting_AllowUnofficialAnswers_RebuildsWordMap(t *testing.T) {
	// This acts as a small integration test toggling the AllowUnofficialAnswers
	// setting and ensuring that it reindexes the remaining answers.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Setup the state with some unofficial answers in it.
	state := NewState(t, "nytbee-20200408.html")
	state.Status = model.StatusSolving
	state.Words = map[string]int{
		"COCONUT":   0,
		"CONCOCT":   1,
		"CONCOCTOR": 2,
		"CONTO":     3,
		"COUNTRY":   11,
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Set the AllowUnofficialAnswers setting to false
	response := Channel.PUT("/setting/allow_unofficial_answers", `false`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		expected := map[string]int{
			"COCONUT": 0,
			"CONCOCT": 1,
			"COUNTRY": 8,
		}
		assert.Equal(t, expected, state.Words)
	})
}

func TestRoute_UpdateSetting_AllowUnofficialAnswers_SendsCompleteEvent(t *testing.T) {
	// This acts as a small integration test toggling the AllowUnofficialAnswers
	// setting and ensuring that when it clears unofficial answers if the puzzle
	// is complete it sends the complete event.
	router, pool, _ := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)

	// AllowUnofficialAnswers in the settings.
	settings := Settings{AllowUnofficialAnswers: true}
	require.NoError(t, SetSettings(conn, Channel.name, settings))

	// Setup the state with some unofficial answers in it.
	state := NewState(t, "nytbee-20200408.html")
	state.Status = model.StatusSolving

	// These are all of the official answers plus an unofficial one.  Once we
	// remove the unofficial answer the puzzle should be solved.
	state.Words = map[string]int{
		"COCONUT":   0,
		"CONCOCT":   1,
		"CONCOCTOR": 2,
		"CONTORT":   4,
		"CONTOUR":   5,
		"COOT":      6,
		"COTTON":    8,
		"COTTONY":   9,
		"COUNT":     10,
		"COUNTRY":   11,
		"COUNTY":    12,
		"COURT":     13,
		"CROUTON":   15,
		"CURT":      18,
		"CUTOUT":    19,
		"NUTTY":     27,
		"ONTO":      29,
		"OUTCRY":    32,
		"OUTRO":     33,
		"OUTRUN":    37,
		"ROOT":      38,
		"ROTO":      40,
		"ROTOR":     41,
		"ROUT":      42,
		"RUNOUT":    43,
		"RUNT":      44,
		"RUNTY":     45,
		"RUTTY":     46,
		"TONY":      49,
		"TOON":      50,
		"TOOT":      51,
		"TORN":      53,
		"TORO":      54,
		"TORT":      57,
		"TOUR":      60,
		"TOUT":      61,
		"TROT":      65,
		"TROUT":     67,
		"TROY":      69,
		"TRYOUT":    70,
		"TURN":      72,
		"TURNOUT":   74,
		"TUTOR":     76,
		"TUTU":      78,
		"TYCOON":    79,
		"TYRO":      80,
		"UNCUT":     81,
		"UNTO":      83,
		"YURT":      85,
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Start listening for events, we should receive 2 a settings event and a
	// state event.
	flush, stop := Channel.SSE("/events", router)
	events := flush()
	assert.Equal(t, 2, len(events))
	assert.Equal(t, "settings", events[0].Kind)
	assert.Equal(t, "state", events[1].Kind)

	// Now toggle the AllowUnofficialAnswers setting to false.
	response := Channel.PUT("/setting/allow_unofficial_answers", `false`, router)
	assert.Equal(t, http.StatusOK, response.Code)

	// We should have received 3 events, a settings event, the state event with
	// the removed answer and a completed event.
	events = stop()
	assert.Equal(t, 3, len(events))
	assert.Equal(t, "settings", events[0].Kind)
	assert.Equal(t, "state", events[1].Kind)
	assert.Equal(t, "complete", events[2].Kind)
}

func TestRoute_UpdateSetting_JSONError(t *testing.T) {
	tests := []struct {
		name    string
		setting string
		json    string
	}{
		{
			name:    "allow_unofficial_answers",
			setting: "allow_unofficial_answers",
			json:    `{`,
		},
		{
			name:    "font_size",
			setting: "font_size",
			json:    `{`,
		},
		{
			name:    "show_answer_placeholders",
			setting: "show_answer_placeholders",
			json:    `{`,
		},
		{
			name:    "invalid setting name",
			setting: "foo_bar_baz",
			json:    `false`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			response := Channel.PUT(fmt.Sprintf("/setting/%s", test.setting), test.json, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_UpdateSetting_LoadSaveError(t *testing.T) {
	tests := []struct {
		name              string
		loadSettingsError error
		saveSettingsError error
		loadStateError    error
		saveStateError    error
	}{
		{
			name:              "error loading settings",
			loadSettingsError: errors.New("forced error"),
		},
		{
			name:              "error saving settings",
			saveSettingsError: errors.New("forced error"),
		},
		{
			name:           "error loading state",
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "nytbee-20200408.html")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.loadSettingsError)
			ForceErrorDuringSettingsSave(t, test.saveSettingsError)
			ForceErrorDuringStateLoad(t, test.loadStateError)
			ForceErrorDuringStateSave(t, test.saveStateError)

			response := Channel.PUT(fmt.Sprintf("/setting/%s", "allow_unofficial_answers"), "false", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_ShuffleLetters(t *testing.T) {
	// This acts as a small integration test toggling the order of the letters of
	// a spelling bee puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Start a puzzle in the selected state.
	state := NewState(t, "nytbee-20200408.html")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Shuffling the letters should fail because the puzzle is not being solved.
	response := Channel.GET("/shuffle", router)
	assert.Equal(t, http.StatusConflict, response.Code)

	// Transition to solving.
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Shuffle the letters
	response = Channel.GET("/shuffle", router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.NotEqual(t, state.Puzzle.Letters, state.Letters)
	})
}

func TestRoute_ShuffleLetters_Error(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  model.Status
		loadStateError error
		saveStateError error
	}{
		{
			name:          "status created",
			initialStatus: model.StatusCreated,
		},
		{
			name:          "status paused",
			initialStatus: model.StatusPaused,
		},
		{
			name:          "status complete",
			initialStatus: model.StatusComplete,
		},
		{
			name:           "error loading state",
			initialStatus:  model.StatusSolving,
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			initialStatus:  model.StatusSolving,
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "nytbee-20200408.html")
			state.Status = test.initialStatus
			require.NoError(t, SetState(conn, Channel.name, state))

			if test.loadStateError != nil {
				ForceErrorDuringStateLoad(t, test.loadStateError)
			}

			if test.saveStateError != nil {
				ForceErrorDuringStateSave(t, test.saveStateError)
			}

			response := Channel.GET("/shuffle", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_ToggleStatus(t *testing.T) {
	// This acts as a small integration test toggling the status of a spelling bee
	// puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Start a puzzle in the selected state.
	state := NewState(t, "nytbee-20200408.html")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Toggle the status, it should transition to solving.
	response := Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
	})

	// Toggle the status again, it should transition to paused. Make sure we
	// sleep for at least a nanosecond first so that the solve was unpaused for
	// a non-zero duration.
	time.Sleep(1 * time.Nanosecond)
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusPaused, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Toggle the status again, it should transition back to solving.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Force the puzzle to be complete.
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	state.Status = model.StatusComplete
	require.NoError(t, SetState(conn, Channel.name, state))

	// Try to toggle the status one more time.  Now that the puzzle is complete
	// it should return a HTTP error.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	state, err = GetState(conn, Channel.name)
	require.NoError(t, err)
	assert.Equal(t, model.StatusComplete, state.Status)
}

func TestRoute_ToggleStatus_Error(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  model.Status
		loadStateError error
		saveStateError error
	}{
		{
			name:          "status created",
			initialStatus: model.StatusCreated,
		},
		{
			name:           "error loading state",
			initialStatus:  model.StatusSelected,
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			initialStatus:  model.StatusSelected,
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "nytbee-20200408.html")
			state.Status = test.initialStatus
			require.NoError(t, SetState(conn, Channel.name, state))

			if test.loadStateError != nil {
				ForceErrorDuringStateLoad(t, test.loadStateError)
			}

			if test.saveStateError != nil {
				ForceErrorDuringStateSave(t, test.saveStateError)
			}

			response := Channel.PUT("/status", "", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_AddAnswer_NoUnofficialAnswers(t *testing.T) {
	// This acts as a small integration test of adding answers to a spelling bee
	// puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	state := NewState(t, "nytbee-20200408.html")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply a correct answer before the puzzle has been started, should get an
	// error.
	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusConflict, response.Code)

	// Transition to solving.
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Now applying the answer should succeed.
	response = Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusCreated, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Contains(t, state.Words, "COCONUT")
	})

	// Applying an incorrect answer should fail.
	response = Channel.POST("/answer", `"CCCC"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Applying an unofficial answer should fail.
	response = Channel.POST("/answer", `"CONCOCTOR"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRoute_AddAnswer_AllowUnofficialAnswers(t *testing.T) {
	// This acts as a small integration test of adding answers to a spelling bee
	// puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	settings := Settings{AllowUnofficialAnswers: true}
	require.NoError(t, SetSettings(conn, Channel.name, settings))

	state := NewState(t, "nytbee-20200408.html")
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Applying an answer from the official list should succeed.
	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusCreated, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Contains(t, state.Words, "COCONUT")
	})

	// Applying an answer from the unofficial list should also succeed.
	response = Channel.POST("/answer", `"CONCOCTOR"`, router)
	assert.Equal(t, http.StatusCreated, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Contains(t, state.Words, "CONCOCTOR")
	})

	// Applying an incorrect answer should fail.
	response = Channel.POST("/answer", `"CCCC"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRoute_AddAnswer_SolvedPuzzleStopsTimer(t *testing.T) {
	// This acts as a small integration test ensuring that the timer stops
	// counting once the puzzle has been solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set the state to have all of the words except for one.
	state := NewState(t, "nytbee-20200408.html")
	state.Status = model.StatusSolving
	state.ApplyAnswer("CONCOCT", false)
	state.ApplyAnswer("CONTORT", false)
	state.ApplyAnswer("CONTOUR", false)
	state.ApplyAnswer("COOT", false)
	state.ApplyAnswer("COTTON", false)
	state.ApplyAnswer("COTTONY", false)
	state.ApplyAnswer("COUNT", false)
	state.ApplyAnswer("COUNTRY", false)
	state.ApplyAnswer("COUNTY", false)
	state.ApplyAnswer("COURT", false)
	state.ApplyAnswer("CROUTON", false)
	state.ApplyAnswer("CURT", false)
	state.ApplyAnswer("CUTOUT", false)
	state.ApplyAnswer("NUTTY", false)
	state.ApplyAnswer("ONTO", false)
	state.ApplyAnswer("OUTCRY", false)
	state.ApplyAnswer("OUTRO", false)
	state.ApplyAnswer("OUTRUN", false)
	state.ApplyAnswer("ROOT", false)
	state.ApplyAnswer("ROTO", false)
	state.ApplyAnswer("ROTOR", false)
	state.ApplyAnswer("ROUT", false)
	state.ApplyAnswer("RUNOUT", false)
	state.ApplyAnswer("RUNT", false)
	state.ApplyAnswer("RUNTY", false)
	state.ApplyAnswer("RUTTY", false)
	state.ApplyAnswer("TONY", false)
	state.ApplyAnswer("TOON", false)
	state.ApplyAnswer("TOOT", false)
	state.ApplyAnswer("TORN", false)
	state.ApplyAnswer("TORO", false)
	state.ApplyAnswer("TORT", false)
	state.ApplyAnswer("TOUR", false)
	state.ApplyAnswer("TOUT", false)
	state.ApplyAnswer("TROT", false)
	state.ApplyAnswer("TROUT", false)
	state.ApplyAnswer("TROY", false)
	state.ApplyAnswer("TRYOUT", false)
	state.ApplyAnswer("TURN", false)
	state.ApplyAnswer("TURNOUT", false)
	state.ApplyAnswer("TUTOR", false)
	state.ApplyAnswer("TUTU", false)
	state.ApplyAnswer("TYCOON", false)
	state.ApplyAnswer("TYRO", false)
	state.ApplyAnswer("UNCUT", false)
	state.ApplyAnswer("UNTO", false)
	state.ApplyAnswer("YURT", false)
	require.NoError(t, SetState(conn, Channel.name, state))
	require.Equal(t, model.StatusSolving, state.Status)

	// Apply the last answer, but wait a bit first to ensure that a non-zero
	// amount of time has passed in the solve.
	time.Sleep(2 * time.Millisecond)

	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusCreated, response.Code)
	VerifyState(t, pool, events, func(state State) {
		require.Equal(t, model.StatusComplete, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0)
	})
}

func TestRoute_AddAnswer_GeniusEvent(t *testing.T) {
	// This acts as a small integration test ensuring that the genius event is
	// emitted once the score crosses the threshold.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set the state to have all of the words except for one.
	state := NewState(t, "nytbee-20200408.html")
	state.Status = model.StatusSolving
	state.ApplyAnswer("CONCOCT", false)
	state.ApplyAnswer("CONTORT", false)
	state.ApplyAnswer("CONTOUR", false)
	state.ApplyAnswer("COOT", false)
	state.ApplyAnswer("COTTON", false)
	state.ApplyAnswer("COTTONY", false)
	state.ApplyAnswer("COUNT", false)
	state.ApplyAnswer("COUNTRY", false)
	state.ApplyAnswer("COUNTY", false)
	state.ApplyAnswer("COURT", false)
	state.ApplyAnswer("CROUTON", false)
	state.ApplyAnswer("CURT", false)
	state.ApplyAnswer("CUTOUT", false)
	state.ApplyAnswer("NUTTY", false)
	state.ApplyAnswer("ONTO", false)
	state.ApplyAnswer("OUTCRY", false)
	state.ApplyAnswer("OUTRO", false)
	state.ApplyAnswer("OUTRUN", false)
	state.ApplyAnswer("ROOT", false)
	state.ApplyAnswer("ROTO", false)
	state.ApplyAnswer("ROTOR", false)
	state.ApplyAnswer("ROUT", false)
	state.ApplyAnswer("RUNOUT", false)
	state.ApplyAnswer("RUNT", false)
	state.ApplyAnswer("RUNTY", false)
	state.ApplyAnswer("RUTTY", false)
	require.NoError(t, SetState(conn, Channel.name, state))
	require.Equal(t, model.StatusSolving, state.Status)

	// The next answer should cross the genius threshold.
	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusCreated, response.Code)
	VerifyGeniusEvent(t, events)
}

func TestRoute_AddAnswer_Error(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int
	}{
		{
			name:     "empty answer",
			json:     `""`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "long answer",
			json:     `"` + RandomString(1023) + `"`,
			expected: http.StatusRequestEntityTooLarge,
		},
		{
			name:     "non-string answer",
			json:     `true`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "nytbee-20200408.html")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			response := Channel.POST("/answer", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_AddAnswer_LoadSaveError(t *testing.T) {
	tests := []struct {
		name              string
		loadSettingsError error
		loadStateError    error
		saveStateError    error
	}{
		{
			name:              "error loading settings",
			loadSettingsError: errors.New("forced error"),
		},
		{
			name:           "error loading state",
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "nytbee-20200408.html")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.loadSettingsError)
			ForceErrorDuringStateLoad(t, test.loadStateError)
			ForceErrorDuringStateSave(t, test.saveStateError)

			response := Channel.POST("/answer", `"COCONUT"`, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_GetEvents(t *testing.T) {
	// This acts as a small integration test ensuring that the event stream
	// receives the events put into a registry.
	router, pool, _ := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)

	// Connect to the stream when there's no puzzle selected, we should receive
	// just the channel's settings.
	_, stop := Channel.SSE("/events", router)
	events := stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "settings", events[0].Kind)

	// Select a puzzle.
	state := NewState(t, "nytbee-20200408.html")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Now reconnect to the stream and we should receive both the settings and the
	// channel's current state.
	flush, stop := Channel.SSE("/events", router)
	events = flush()
	assert.Equal(t, 2, len(events))
	assert.Equal(t, "settings", events[0].Kind)
	assert.Equal(t, "state", events[1].Kind)

	// Toggle the status to solving, this should cause the state to be sent again.
	response := Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Now specify an answer, this should cause the state to be sent again.
	response = Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusCreated, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Now change a setting, this should cause the settings to be sent again.
	response = Channel.PUT("/setting/font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "settings", events[0].Kind)

	// Disconnect, there shouldn't be any events anymore.
	events = stop()
	assert.Equal(t, 0, len(events))
}

func TestRoute_GetEvents_LoadSaveError(t *testing.T) {
	tests := []struct {
		name                   string
		forceSettingsLoadError error
		forceStateLoadError    error
	}{
		{
			name:                   "error loading settings",
			forceSettingsLoadError: errors.New("forced error"),
		},
		{
			name:                "error loading state",
			forceStateLoadError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "nytbee-20200408.html")
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.forceSettingsLoadError)
			ForceErrorDuringStateLoad(t, test.forceStateLoadError)

			// This won't start a background goroutine to send events because the
			// request will fail before reaching that part of the code.
			response := Channel.GET("/events", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_GetAvailableDates(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			name:   "new_york_times",
			source: "new_york_times",
			expected: []string{
				NYTBeeFirstPuzzleDate.Format("2006-01-02"),
				"2018-08-01",
				"2018-09-01",
				"2018-10-01",
				"2018-11-01",
				"2018-12-01",
				"2019-01-01",
				"2019-02-01",
				"2019-03-01",
				"2019-04-01",
				"2019-05-01",
				"2019-06-01",
				"2019-07-01",
				"2019-08-01",
				"2019-09-01",
				"2019-10-01",
				"2019-11-01",
				"2019-12-01",
				time.Now().UTC().Format("2006-01-02"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			response := GET("/spellingbee/dates", router)
			assert.Equal(t, http.StatusOK, response.Code)

			var dates map[string][]string
			require.NoError(t, render.DecodeJSON(response.Result().Body, &dates))

			for _, expected := range test.expected {
				index := sort.SearchStrings(dates[test.source], expected)
				require.True(t, index != len(dates))
				assert.Equal(t, expected, dates[test.source][index])
			}
		})
	}
}

// VerifySettings performs test specific verifications on the settings objects
// in both event and database forms.
func VerifySettings(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s Settings)) {
	t.Helper()

	// First check that we've received a single settings event with the correct
	// values
	found := Events(events, "settings")
	require.Equal(t, 1, len(found), "incorrect number of events found")

	settings := found[0].Payload.(Settings)
	fn(settings)

	// Next check that the database has a valid settings object
	conn := NewRedisConnection(t, pool)
	settings, err := GetSettings(conn, Channel.name)
	require.NoError(t, err)
	fn(settings)
}

// VerifyState performs common verifications for state objects in both event
// and database forms and then calls a custom verify function for test specific
// verifications.
func VerifyState(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(State)) {
	t.Helper()

	// First check that we've received a single state event that has the correct
	// values
	found := Events(events, "state")
	require.Equal(t, 1, len(found), "incorrect number of events found")

	// Event should never have the answers
	state := found[0].Payload.(State)
	assert.Nil(t, state.Puzzle.OfficialAnswers)
	assert.Nil(t, state.Puzzle.UnofficialAnswers)
	fn(state)

	// Next check that the database has a valid state object
	state, err := GetState(NewRedisConnection(t, pool), Channel.name)
	require.NoError(t, err)

	// Database should always have the answers
	assert.NotNil(t, state.Puzzle.OfficialAnswers)
	assert.NotNil(t, state.Puzzle.UnofficialAnswers)
	fn(state)
}

// VerifyGeniusEvent ensures that a genius event is present in the event stream.
func VerifyGeniusEvent(t *testing.T, events <-chan pubsub.Event) {
	t.Helper()

	found := Events(events, "genius")
	require.Equal(t, 1, len(found))
}

// Events extracts events of a particular kind from a channel.  It consumes all
// events in the channel that are available at the time of the call.
func Events(events <-chan pubsub.Event, kind string) []pubsub.Event {
	var found []pubsub.Event

	for done := false; !done; {
		select {
		case event := <-events:
			if event.Kind != kind {
				continue
			}

			found = append(found, event)

		default:
			done = true
		}
	}

	return found
}

// GET performs a HTTP GET request to the provided router.
func GET(url string, router chi.Router) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

// ChannelClient is a client that makes requests against the URL of a particular
// user's channel.
type ChannelClient struct {
	name string
}

func (c ChannelClient) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/spellingbee", c.name, url)
	return GET(url, router)
}

func (c ChannelClient) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/spellingbee", c.name, url)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
	router.ServeHTTP(recorder, request)
	return recorder
}

func (c ChannelClient) POST(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/spellingbee", c.name, url)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	router.ServeHTTP(recorder, request)
	return recorder
}

// SSE performs a streaming request to the provided router.  Because the router
// won't immediately return, this request is done in a background goroutine.
// When the main thread wishes to read events that have been received thus far
// the flush method can be called and it will return any queued up events.  When
// the main thread wishes to close the connection to the router the stop method
// can be called and it will return any unread events.
func (c ChannelClient) SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join("/spellingbee", c.name, url)
	recorder := CreateTestResponseRecorder()

	flush = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		reader, err := recorder.Body()
		if err != nil {
			return nil
		}

		var events []pubsub.Event
		for {
			bs, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			if !bytes.HasPrefix(bs, []byte("data:")) {
				continue
			}

			var event pubsub.Event
			json.Unmarshal(bs[5:], &event)
			events = append(events, event)
		}

		return events
	}

	stop = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		recorder.Close()
		return flush()
	}

	request := httptest.NewRequest(http.MethodGet, url, nil)
	go router.ServeHTTP(recorder, request)

	return flush, stop
}

func RandomString(n int) string {
	var alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	bs := make([]rune, n)
	for i := range bs {
		bs[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(bs)
}

// Create a http.ResponseWriter that synchronizes whenever reads or writes
// happen so that there are no races in a multiple goroutine environment.
// Additionally implement the http.CloseNotifier interface so that requests can
// be stopped by tests.
type TestResponseRecorder struct {
	sync.Mutex
	headers http.Header
	body    *bytes.Buffer
	close   chan bool
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		close:   make(chan bool, 1),
	}
}

func (r *TestResponseRecorder) Header() http.Header {
	r.Lock()
	defer r.Unlock()

	return r.headers
}

func (r *TestResponseRecorder) Write(bs []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.body.Write(bs)
}

func (r *TestResponseRecorder) Body() (*bufio.Reader, error) {
	r.Lock()
	defer r.Unlock()

	bs, err := ioutil.ReadAll(r.body)
	if err != nil {
		return nil, err
	}
	r.body.Reset()
	return bufio.NewReader(bytes.NewReader(bs)), nil
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	r.Lock()
	defer r.Unlock()

	return r.close
}

func (r *TestResponseRecorder) Close() {
	r.Lock()
	defer r.Unlock()

	r.close <- true
}

func (r *TestResponseRecorder) WriteHeader(int) {
	// Not used
}
