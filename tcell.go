package main

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func initTcell() (s tcell.Screen, err error) {
	if s, err = tcell.NewScreen(); err != nil {
		return
	}
	if err = s.Init(); err != nil {
		return
	}
	defStyle := tcell.StyleDefault
	defStyle = defStyle.Background(tcell.ColorBlack)
	defStyle = defStyle.Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)
	s.Clear()
	return
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func handleTcellEvents(screen tcell.Screen, events chan Event) {
	for {
		switch ev := screen.PollEvent().(type) {
		case *tcell.EventResize:
			screen.Sync()
			events <- redrawEvent
		case *tcell.EventKey:
			switch key := ev.Key(); key {
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
	}
}
