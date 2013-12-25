package main

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
)

var matcher = regexp.MustCompile(`^(?:<(\d+)>)?(.*)`)

type SyslogMsg struct {
	Priority int
	Facility int
	Severity int
	Msg      string
}

func parseSyslogMsg(buf []byte) (*SyslogMsg, error) {
	matches := matcher.FindSubmatch(buf)
	if len(matches) == 0 {
		return nil, errors.New("No match")
	}

	var prio int
	var err error
	if len(matches[2]) != 0 {
		prio, err = strconv.Atoi(string(matches[1]))
		if err != nil {
			return nil, errors.New("prio failed to convert")
		}
	} else {
		// default prio a relay must write if none is readable
		prio = 13
	}

	m := &SyslogMsg{
		Priority: prio,
		Facility: prio / 8,
		Severity: prio % 8}
	m.Msg = string(bytes.TrimSpace(matches[2]))
	return m, nil
}
