import React from "react";
import {DateChooser, Switch} from "common/nav";
import formatISO from "date-fns/formatISO";
import parseISO from "date-fns/parseISO";

export function ViewsDropdown(props) {
  const base = `${document.location.origin}/${props.channel}/acrostic`;
  const copyToClipboard = (id) => {
    const input = document.getElementById(id);
    if (input !== null) {
      input.select();
      document.execCommand("copy");
      document.getSelection().empty();
    }
  };

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Views
      </button>
      <div id="views-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="views-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">Streamer View</div>
            <div>
              <small className="text-muted">
                This link allows changing of settings as well as the active
                puzzle. Only share it with those you fully trust.
              </small>
            </div>
            <div className="input-group">
              <input id="streamer-url" type="text" className="form-control" value={`${base}/streamer`} readOnly/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={() => copyToClipboard("streamer-url")}>Copy</button>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Progress View</div>
            <div>
              <small className="text-muted">
                This link will only show the progress of the puzzle solve. As
                cells are populated with values they will change color, but the
                value put in the cells will not be visible. This link is
                intended to be shared with others who are solving the puzzle at
                the same time, but shouldn't see any answers.
              </small>
            </div>
            <div className="input-group">
              <input id="progress-url" type="text" className="form-control" value={`${base}/progress`} readOnly/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={() => copyToClipboard("progress-url")}>Copy</button>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Participant View</div>
            <div>
              <small className="text-muted">
                This link will show the complete progress of the solve as it
                happens in real-time. This link is safe to share with anyone.
              </small>
            </div>
            <div className="input-group">
              <input id="participant-url" type="text" className="form-control" value={`${base}`} readOnly/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={() => copyToClipboard("participant-url")}>Copy</button>
              </div>
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}

export function SettingsDropdown(props) {
  const settings = props.settings;

  // Update a setting with its new value.
  function update(name, value) {
    return function() {
      if (settings[name] === value) {
        return;
      }

      fetch(`/api/acrostic/${props.channel}/setting/${name}`,
        {
          method: "PUT",
          body: JSON.stringify(value),
        });
    };
  }

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Settings
      </button>
      <div id="settings-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="settings-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">Only correct answers</div>
            <div>
              <small className="text-muted">
                This setting enables the behavior that only correct answers or
                correct partial answers will be accepted. With this enabled a
                cell in the puzzle can never contain an incorrect value.
              </small>
            </div>
            <Switch checked={settings.only_allow_correct_answers} onClick={update("only_allow_correct_answers", !settings.only_allow_correct_answers)}/>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Clue font size</div>
            <div>
              <small className="text-muted">
                This setting allows the size of the font used to render clues to
                be adjusted.
              </small>
            </div>
            <div className="btn-group" role="group">
              <button type="button" className={settings.clue_font_size === "normal" ? "btn btn-success" : "btn btn-dark"} onClick={update("clue_font_size", "normal")}>Normal</button>
              <button type="button" className={settings.clue_font_size === "large" ? "btn btn-success" : "btn btn-dark"} onClick={update("clue_font_size", "large")}>Large</button>
              <button type="button" className={settings.clue_font_size === "xlarge" ? "btn btn-success" : "btn btn-dark"} onClick={update("clue_font_size", "xlarge")}>Extra Large</button>
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}

export function PuzzleDropdown({channel, setErrorMessage}) {
  const [minNYTDate, setMinNYTDate] = React.useState(null);
  const [nytDates, setNYTDates] = React.useState(new Set());

  // Select a puzzle for the channel.  If the puzzle fails to load properly
  // then a simple error message will be displayed until a page reload or a
  // successful puzzle load.
  const setPuzzle = (payload) => {
    return fetch(`/api/acrostic/${channel}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      })
      .then(response => {
        if (!response.ok) {
          throw new Error("Unable to load puzzle.")
        }

        setErrorMessage(null);
      })
      .then(() => {
        // Hide the dropdown menu after a selection is successfully made.
        const menu = document.getElementById("puzzle-dropdown-menu");
        if (menu) {
          menu.classList.remove("show");
        }
      })
      .catch(error => setErrorMessage(error.message));
  };

  // Determine if a puzzle is available for a particular date.
  const isNYTPuzzleAvailableForDate = (date) => {
    const iso = formatISO(date, {representation: "date"});
    return nytDates.has(iso);
  };

  // Select a NYT puzzle for a specific date.
  const onNYTDateSelected = (date) => {
    if (!date) {
      return;
    }

    date = formatISO(date, {representation: "date"});
    return setPuzzle({"new_york_times_date": date});
  };

  React.useEffect(() => {
    fetch(`/api/acrostic/dates/nytimes`)
      .then(response => {
        if (!response.ok) {
          throw new Error("Unable to load available puzzle dates.");
        }

        return response.json();
      })
      .then(dates => {
        setMinNYTDate(parseISO(dates[0]));
        setNYTDates(new Set(dates));
      })
      .catch(error => setErrorMessage(error.message));
  }, [setErrorMessage, setMinNYTDate, setNYTDates]);

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Puzzle
      </button>
      <div id="puzzle-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="puzzle-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">New York Times</div>
            <div>
              <small className="text-muted">
                Select a date to solve that day's puzzle from the archives of
                the New York Times.
              </small>
            </div>
            <div className="input-group">
              <DateChooser
                onClick={onNYTDateSelected}
                filterDate={isNYTPuzzleAvailableForDate}
                minDate={minNYTDate}
              />
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}