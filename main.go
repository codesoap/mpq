package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"os"
	"time"
)

const mpdAddr = "localhost:6600"

type state struct {
	mpdState mpdState // play, stop or pause
	// elapsed time of the current song in seconds;
	// nil when stopped:
	elapsed *float32
	// songid of the currently playing or paused song;
	// nil when stopped:
	songID      *int
	highlighted int // index of the highlighted song in the queue
	queue       []song
}

type song struct {
	uri      string // the URI/file; only used as fallback title
	songID   int
	duration float32
	artist   string // "" if unknown
	title    string // "" if unknown
	album    string // "" if unknown
	track    *int   // the track number within the album; nil if unknown
}

type event int

const (
	updateStateEvent event = iota
	playHighlightedEvent
	togglePauseEvent
	deleteHighlightedEvent
	highlightPrevEvent
	highlightNextEvent
	seekBackwardsEvent
	seekForwardsEvent
	movePrevEvent
	moveNextEvent
	redrawEvent
	quitEvent
)

func main() {
	screen, err := initTcell()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize tcell: %v\n", err)
		os.Exit(1)
	}

	state, err := getState()
	if err != nil {
		screen.Fini()
		fmt.Fprintf(os.Stderr, "Could not get mpd state: %v\n", err)
		os.Exit(1)
	}

	events := make(chan event)
	go handleTcellEvents(screen, events)
	go handleMpdEvents(events)
	if err = runEventLoop(state, screen, events); err != nil {
		screen.Fini()
		fmt.Fprintf(os.Stderr, "Error during execution: %v\n", err)
		os.Exit(1)
	}
	screen.Fini()
}

func runEventLoop(state state, screen tcell.Screen, events chan event) error {
	var err error
	updateElapsedTicker := time.NewTicker(time.Second)
	for {
		select {
		case event := <-events:
			switch event {
			case updateStateEvent:
				oldHighlighted := state.highlighted
				if state, err = getState(); err != nil {
					return fmt.Errorf("could not get mpd state: %v", err)
				}
				state.highlighted = oldHighlighted
				if state.highlighted < 0 {
					state.highlighted = 0
				}
				if state.highlighted >= len(state.queue) {
					state.highlighted = len(state.queue) - 1
				}
				draw(state, screen)
			case playHighlightedEvent:
				if err = playHighlighted(state); err != nil {
					return err
				}
			case togglePauseEvent:
				if err = togglePause(state); err != nil {
					return err
				}
			case deleteHighlightedEvent:
				if err = deleteHighlighted(state); err != nil {
					return err
				}
			case highlightPrevEvent:
				if state.highlighted > 0 {
					state.highlighted -= 1
					draw(state, screen)
				}
			case highlightNextEvent:
				if state.highlighted < len(state.queue)-1 {
					state.highlighted += 1
					draw(state, screen)
				}
			case movePrevEvent:
				if err = moveHighlightedUpwards(&state); err != nil {
					return err
				}
			case moveNextEvent:
				if err = moveHighlightedDownwards(&state); err != nil {
					return err
				}
			case seekBackwardsEvent:
				if state.mpdState != stopMPDState {
					if err = seekBackwards(5); err != nil {
						return err
					}
				}
			case seekForwardsEvent:
				if state.mpdState != stopMPDState {
					if err = seekForwards(5); err != nil {
						return err
					}
				}
			case redrawEvent:
				draw(state, screen)
			case quitEvent:
				return nil
			}
		case <-updateElapsedTicker.C:
			if updated := updateElapsed(&state); updated {
				draw(state, screen)
			}
		}
	}
}

func updateElapsed(state *state) bool {
	if state.mpdState == playMPDState {
		*state.elapsed += 1
		return true
	}
	return false
}

/* How it could look:
Paused at 01:18 / 02:15
  03:43 [#4 of Meddle] Pink Floyd - San Tropez
> 02:15 [#5 of Meddle] Pink Floyd - Seamus
  23:35 [#6 of Meddle] Pink Floyd - Echoes
*/
func draw(state state, screen tcell.Screen) {
	screen.Clear()
	emitStr(screen, 0, 0, tcell.StyleDefault, getTopbar(state))

	cropQueueIfTooLong(&state, screen)
	for i, s := range state.queue {
		if state.songID != nil && *state.songID == s.songID {
			emitStr(screen, 0, i+1, tcell.StyleDefault, ">")
		}
		var line string
		if s.title == "" {
			line = fmt.Sprintf("%s", s.uri)
		} else if s.artist == "" {
			line = fmt.Sprintf("%s", s.title)
		} else if s.track == nil || s.album == "" {
			line = fmt.Sprintf("%s - %s", s.artist, s.title)
		} else {
			line = fmt.Sprintf("[#%02d of %s] %s - %s", *s.track, s.album, s.artist, s.title)
		}
		dur := fmt.Sprintf("%02d:%02d ", int(s.duration)/60, int(s.duration)%60)
		if i == state.highlighted {
			emitStr(screen, 2, i+1, tcell.StyleDefault.Reverse(true), dur+line)
		} else {
			emitStr(screen, 2, i+1, tcell.StyleDefault, dur+line)
		}
	}
	screen.Show()
}

func getTopbar(state state) string {
	switch state.mpdState {
	case playMPDState:
		currentSong, err := getCurrentSong(state)
		if err != nil {
			return "playing"
		}
		return fmt.Sprintf("playing at %02d:%02d / %02d:%02d",
			int(*state.elapsed)/60,
			int(*state.elapsed)%60,
			int(currentSong.duration)/60,
			int(currentSong.duration)%60)
	case pauseMPDState:
		currentSong, err := getCurrentSong(state)
		if err != nil {
			return "paused"
		}
		return fmt.Sprintf("paused at %02d:%02d / %02d:%02d",
			int(*state.elapsed)/60,
			int(*state.elapsed)%60,
			int(currentSong.duration)/60,
			int(currentSong.duration)%60)
	}
	return "stopped"
}

// cropQueueIfTooLong will remove songs from state.queue, if the screen
// is not high enough to fit all of them. The function ensures that the
// highlighted song is alway visible.
//
// state.highlighted also will be adapted to point to the same song
// after cropping.
func cropQueueIfTooLong(state *state, screen tcell.Screen) {
	if _, h := screen.Size(); h <= len(state.queue) {
		newQueueLen := h - 1
		start := state.highlighted - ((newQueueLen - 1) / 2)
		end := start + newQueueLen
		if start < 0 {
			start = 0
			end = newQueueLen
		} else if end > len(state.queue) {
			end = len(state.queue)
			start = end - newQueueLen
		}
		state.queue = state.queue[start:end]
		state.highlighted -= start
	}
}

func getCurrentSong(state state) (song, error) {
	for _, s := range state.queue {
		if s.songID == *state.songID {
			return s, nil
		}
	}
	return song{}, fmt.Errorf("currently playing song not in queue")
}
