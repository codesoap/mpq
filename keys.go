package main

import "github.com/gdamore/tcell"

const keyBindingInfo = `Key bindings:
q          : quit
enter      : play highlighted song
space      : toggle play/pause
up,k       : highlight previous song
down,j     : highlight next song
alt+up/k   : move highlighted song up
alt+down/j : move highlighted song down
left,h     : seek backwards 5s
right,l    : seek forwards 5s
d          : remove song from queue
c          : clear queue`

func handleKeyEvents(ev *tcell.EventKey, events chan event) {
	switch ev.Key() {
	case tcell.KeyEnter:
		events <- playHighlightedEvent
	case tcell.KeyUp:
		handlePrevKey(ev, events)
	case tcell.KeyDown:
		handleNextKey(ev, events)
	case tcell.KeyLeft:
		events <- seekBackwardsEvent
	case tcell.KeyRight:
		events <- seekForwardsEvent
	case tcell.KeyRune:
		// "Normal" keys are handled here.
		switch ev.Rune() {
		case ' ':
			events <- togglePauseEvent
		case 'h':
			events <- seekBackwardsEvent
		case 'j':
			handleNextKey(ev, events)
		case 'k':
			handlePrevKey(ev, events)
		case 'l':
			events <- seekForwardsEvent
		case 'd':
			events <- deleteHighlightedEvent
		case 'c':
			events <- clearEvent
		case 'q':
			events <- quitEvent
		}
	}
}

func handlePrevKey(ev *tcell.EventKey, events chan event) {
	if ev.Modifiers()&tcell.ModAlt > 0 {
		events <- movePrevEvent
	} else {
		events <- highlightPrevEvent
	}
}

func handleNextKey(ev *tcell.EventKey, events chan event) {
	if ev.Modifiers()&tcell.ModAlt > 0 {
		events <- moveNextEvent
	} else {
		events <- highlightNextEvent
	}
}
