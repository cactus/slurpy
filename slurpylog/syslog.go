package slurpylog

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"github.com/cactus/gologit"
)

var matcher = regexp.MustCompile(`^(?:<(\d+)>)?(.*)`)

var severityMap = map[int]string {
	0: "EMERG",
	1: "ALERT",
	2: "CRIT",
	3: "ERR",
	4: "WARNING",
	5: "NOTICE",
	6: "INFO",
	7: "DEBUG",
}

var facilityMap = map[int]string {
	0:   "KERN",
	8:   "USER",
	16:  "MAIL",
	24:  "DAEMON",
	32:  "AUTH",
	40:  "SYSLOG",
	48:  "LPR",
	56:  "NEWS",
	64:  "UUCP",
	72:  "CRON",
	80:  "AUTHPRIV",
	88:  "FTP",
	128: "LOCAL0",
	136: "LOCAL1",
	144: "LOCAL2",
	152: "LOCAL3",
	160: "LOCAL4",
	168: "LOCAL5",
	176: "LOCAL6",
	184: "LOCAL7",
}

func FacilityGetName(facility int) (string, error) {
	name, ok := facilityMap[facility]
	if !ok {
		return "", errors.New("Out of range")
	}
	return name, nil
}

func SeverityGetName(severity int) (string, error) {
	name, ok := severityMap[severity]
	if !ok {
		return "", errors.New("Out of range")
	}
	return name, nil
}

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
		Severity: prio % 8,
		Facility: prio - (prio % 8)}
	m.Msg = string(bytes.TrimSpace(matches[2]))
	return m, nil
}
