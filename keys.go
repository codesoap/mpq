package main

import "github.com/gdamore/tcell"

const keyBindingInfo = `Key bindings:
q       : quit
enter   : play highlighted song
space   : toggle play/pause
up      : highlight previous song
down    : highlight next song
alt-up  : move highlighted song up
alt-down: move highlighted song down
left    : seek backwards 5s
right   : seek forwards 5s
d       : remove song from queue`

func handleKeyEvents(ev *tcell.EventKey, events chan event) {
	switch ev.Key() {
	case tcell.KeyEnter:
		events <- playHighlightedEvent
	case tcell.KeyUp:
		if ev.Modifiers()&tcell.ModAlt > 0 {
			events <- movePrevEvent
		} else {
			events <- highlightPrevEvent
		}
	case tcell.KeyDown:
		if ev.Modifiers()&tcell.ModAlt > 0 {
			events <- moveNextEvent
		} else {
			events <- highlightNextEvent
		}
	case tcell.KeyLeft:
		events <- seekBackwardsEvent
	case tcell.KeyRight:
		events <- seekForwardsEvent
	case tcell.KeyRune:
		// "Normal" keys are handled here.
		switch ev.Rune() {
		case ' ':
			events <- togglePauseEvent
		case 'd':
			events <- deleteHighlightedEvent
		case 'q':
			events <- quitEvent
		}
	}
}
