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

func handleTcellEvents(screen tcell.Screen, events chan event) {
	for {
		switch ev := screen.PollEvent().(type) {
		case *tcell.EventResize:
			screen.Sync()
			events <- redrawEvent
		case *tcell.EventKey:
			handleKeyEvents(ev, events)
		}
	}
}
