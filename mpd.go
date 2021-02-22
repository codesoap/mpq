package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// TODO: Timeouts for network stuff.
// FIXME: Should I look for tags case-insensitively? The documentation
//        doesn't say anything about this.

type mpdState int

const (
	playMPDState mpdState = iota
	stopMPDState
	pauseMPDState
)

func handleMpdEvents(events chan event) {
	for {
		_, err := executeMPDCommand("idle playlist player")
		if err != nil {
			return
		}
		events <- updateStateEvent
	}
}

func getState() (state state, err error) {
	status, err := executeMPDCommand("status")
	if err != nil {
		return
	}
	if state.mpdState, err = getMPDState(status); err != nil {
		return
	}
	if state.elapsed, err = getElapsed(status); err != nil {
		return
	}
	if state.songID, err = getSongID(status); err != nil {
		return
	}
	if err = fillQueue(&state); err != nil {
		return
	}
	return
}

func fillQueue(state *state) error {
	info, err := executeMPDCommand("playlistinfo")
	if err != nil {
		return err
	}
	var s song
	var file, title, artist, album string
	var track *int
	for _, line := range strings.Split(info, "\n") {
		split := strings.SplitN(line, ": ", 2)
		switch split[0] {
		case "file":
			if len(file) > 0 {
				s.displayName = composeDisplayName(file, title, artist, album, track)
				state.queue = append(state.queue, s)

				// Reset before parsing the next song:
				file = ""
				title = ""
				artist = ""
				album = ""
				track = nil
			}
			if len(split) < 2 {
				return fmt.Errorf("encountered empty URI")
			}
			file = split[1]
		case "Id":
			if len(split) > 1 {
				if s.songID, err = strconv.Atoi(split[1]); err != nil {
					return fmt.Errorf("could not parse songid: %s", err.Error())
				}
			}
		case "duration":
			if len(split) > 1 {
				f, err := strconv.ParseFloat(split[1], 32)
				if err != nil {
					return fmt.Errorf("could not parse duration: %s", err.Error())
				}
				s.duration = float32(f)
			}
		case "Title":
			if len(split) > 1 {
				title = split[1]
			}
		case "Artist":
			if len(split) > 1 {
				artist = split[1]
			}
		case "Album":
			if len(split) > 1 {
				album = split[1]
			}
		case "Track":
			if len(split) > 1 {
				i, err := strconv.Atoi(split[1])
				if err != nil {
					return fmt.Errorf("could not parse track: %s", err.Error())
				}
				track = &i
			}
		}
	}
	if len(file) > 0 {
		s.displayName = composeDisplayName(file, title, artist, album, track)
		state.queue = append(state.queue, s) // add last song to queue
	}
	return nil
}

func composeDisplayName(file, title, artist, album string, track *int) string {
	if title == "" {
		return fmt.Sprintf("%s", file)
	} else if artist == "" {
		return fmt.Sprintf("%s", title)
	} else if track == nil || album == "" {
		return fmt.Sprintf("%s - %s", artist, title)
	}
	return fmt.Sprintf("[#%02d of %s] %s - %s", *track, album, artist, title)
}

func getMPDState(status string) (mpdState mpdState, err error) {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "state: ") && len(line) > 7 {
			switch line[7:] {
			case "play":
				return playMPDState, nil
			case "stop":
				return stopMPDState, nil
			case "pause":
				return pauseMPDState, nil
			}
		}
	}
	err = fmt.Errorf("mpdState not found")
	return
}

func getElapsed(status string) (*float32, error) {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "elapsed: ") && len(line) > 9 {
			s, err := strconv.ParseFloat(line[9:], 32)
			s32 := float32(s)
			return &s32, err
		}
	}
	return nil, nil
}

func getSongID(status string) (*int, error) {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "songid: ") && len(line) > 8 {
			i, err := strconv.Atoi(line[8:])
			return &i, err
		}
	}
	return nil, nil
}

func executeMPDCommand(command string) (resp string, err error) {
	conn, err := initiateMPDConnection()
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		return
	}
	fmt.Fprintf(conn, "%s\n", command)
	var respBuilder strings.Builder
	var line string
	connReader := bufio.NewReader(conn)
	for {
		if line, err = connReader.ReadString('\n'); err != nil {
			return
		}
		if line == "OK\n" {
			break
		} else if strings.HasPrefix(line, "ACK ") {
			msg := "received mpd error '%s' while executing '%s'"
			err = fmt.Errorf(msg, strings.TrimSpace(line), command)
			break
		} else {
			respBuilder.WriteString(line)
		}
	}
	return respBuilder.String(), err
}

func initiateMPDConnection() (conn *net.TCPConn, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", mpdAddr)
	if err != nil {
		return
	}
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return
	}
	var line string
	connReader := bufio.NewReader(conn)
	if line, err = connReader.ReadString('\n'); err != nil {
		return
	}
	if !strings.HasPrefix(line, "OK MPD ") {
		err = fmt.Errorf("no mpd server found")
	}
	return
}

func playHighlighted(state state) error {
	if len(state.queue) == 0 {
		return nil
	}
	song := state.queue[state.highlighted]
	_, err := executeMPDCommand(fmt.Sprintf("playid %d", song.songID))
	return err
}

func togglePause(state state) error {
	switch state.mpdState {
	case playMPDState:
		_, err := executeMPDCommand("pause 1")
		return err
	case pauseMPDState:
		_, err := executeMPDCommand("pause 0")
		return err
	}
	return nil
}

func deleteHighlighted(state state) error {
	if len(state.queue) == 0 {
		return nil
	}
	song := state.queue[state.highlighted]
	_, err := executeMPDCommand(fmt.Sprintf("deleteid %d", song.songID))
	if err != nil && strings.Contains(err.Error(), "No such song") {
		// This usually happens when pressing the delete button too quickly.
		return nil
	}
	return err
}

func clear(state state) error {
	_, err := executeMPDCommand("clear")
	return err
}

func moveHighlightedUpwards(state *state) error {
	if len(state.queue) == 0 {
		return nil
	}
	if state.highlighted == 0 {
		// Can't move over the top. Just ignore.
		return nil
	}
	cmd := fmt.Sprintf("move %d %d", state.highlighted, state.highlighted-1)
	state.highlighted -= 1
	_, err := executeMPDCommand(cmd)
	return err
}

func moveHighlightedDownwards(state *state) error {
	if len(state.queue) == 0 {
		return nil
	}
	if state.highlighted >= len(state.queue)-1 {
		// Can't move below the bottom. Just ignore.
		return nil
	}
	cmd := fmt.Sprintf("move %d %d", state.highlighted, state.highlighted+1)
	state.highlighted += 1
	_, err := executeMPDCommand(cmd)
	return err
}

func seekBackwards(state state, seconds int) error {
	if state.mpdState == stopMPDState {
		return nil
	}
	_, err := executeMPDCommand(fmt.Sprintf("seekcur -%d", seconds))
	return err
}

func seekForwards(state state, seconds int) error {
	if state.mpdState == stopMPDState {
		return nil
	}

	// Some logic is required here, because mpd behaves weirdly, when
	// trying to seek across the end of a song:
	song, err := getCurrentSong(state)
	if err != nil {
		return err
	}
	remaining := song.duration - *state.elapsed
	if remaining <= 0.4 {
		// No need to seek, if the end is almost reached.
		return nil
	} else if remaining-0.4 < float32(seconds) {
		_, err = executeMPDCommand(fmt.Sprintf("seekcur +%f", remaining-0.3))
	} else {
		_, err = executeMPDCommand(fmt.Sprintf("seekcur +%d", seconds))
	}

	if err != nil && strings.Contains(err.Error(), "Decoder failed to seek") {
		// This seems to happen, when trying to seek shortly after the song
		// changed.
		return nil
	}
	return err
}
