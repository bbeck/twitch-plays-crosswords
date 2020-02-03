import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "./nav.css";

export function Nav(props) {
  // If we're not on the streamer view then don't show the controls that can
  // modify the state of the application.
  if (!props.view || props.view !== "streamer") {
    return (
      <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
        <div className="navbar-brand">Twitch Plays Crosswords</div>
      </nav>
    );
  }

  return (
    <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
      <div className="navbar-brand">Twitch Plays Crosswords</div>

      <ul className="navbar-nav ml-auto">
        <StartPauseButton channel={props.channel} status={props.status}/>
        <ViewsDropdown channel={props.channel}/>
        <SettingsDropdown channel={props.channel} settings={props.settings}/>
        <PuzzleDropdown channel={props.channel}/>
      </ul>
    </nav>
  );
}

function StartPauseButton(props) {
  const status = props.status;

  let message;
  if (status === "created") {
    message = "Start";
  } else if (status === "paused") {
    message = "Unpause";
  } else if (status === "solving") {
    message = "Pause";
  } else if (status === undefined || status === "complete") {
    return null;
  }

  // Toggle the status.
  function toggle() {
    fetch(`/api/${props.channel}/crossword/status`,
      {
        method: "PUT",
      });
  }

  return (
    <form className="form-inline nav-item">
      <button className="btn btn-success" type="button" onClick={toggle}>{message}</button>
    </form>
  );
}

function ViewsDropdown(props) {
  const base = `${document.location.origin}/${props.channel}`;
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

function SettingsDropdown(props) {
  const settings = props.settings;

  // Update a setting with its new value.
  function update(name, value) {
    return function() {
      if (settings[name] === value) {
        return;
      }

      fetch(`/api/${props.channel}/crossword/setting/${name}`,
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
            <div className="lead">Reduced set of clues</div>
            <div>
              <small className="text-muted">
                This setting enables some of the clues to be hidden thus making
                the puzzle significantly more difficult to solve.
              </small>
            </div>
            <div className="btn-group" role="group">
              <button type="button" className={settings.clues_to_show === "all" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "all")}>Show All</button>
              <button type="button" className={settings.clues_to_show === "down" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "down")}>Hide Across</button>
              <button type="button" className={settings.clues_to_show === "across" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "across")}>Hide Down</button>
              <button type="button" className={settings.clues_to_show === "none" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "none")}>Hide All</button>
            </div>
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

function PuzzleDropdown(props) {
  // Set the date of the puzzle to use
  const onSetPuzzleDate = (e) => {
    const elem = document.getElementById("puzzle-date-input");
    const date = elem.value;

    fetch(`/api/${props.channel}/crossword/date`,
      {
        method: "PUT",
        body: JSON.stringify({"date": date}),
      });
  };

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
              <input id="puzzle-date-input" type="date" className="form-control"/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={onSetPuzzleDate}>Load
                </button>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Download a .puz file</div>
            <div>
              <small className="text-muted">
                Input a URL to a .puz file to solve an existing puzzle hosted
                somewhere online.
              </small>
            </div>
            <div className="input-group">
              <input type="url" className="form-control"/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark">Load</button>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Upload a .puz file</div>
            <div>
              <small className="text-muted">
                Upload your own puzzle file in .puz format. More info about
                the .puz file format can be found&nbsp;
                <a href="http://fileformats.archiveteam.org/wiki/PUZ_(crossword_puzzles)">here</a>.
              </small>
            </div>
            <div className="input-group">
              <button type="button" className="btn btn-dark">Choose file</button>
              <input type="file" className="d-none" accept=".puz"/>
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}

function Switch(props) {
  return (
    <label className="switch">
      <input type="checkbox" className="form-check-input success" checked={props.checked} readOnly/>
      <span className="slider round" onClick={props.onClick}/>
    </label>
  );
}
