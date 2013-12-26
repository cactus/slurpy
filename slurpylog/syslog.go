package slurpylog

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
)

var matcher = regexp.MustCompile(`^(?:<(\d+)>)?(.*)`)

var facilityMap = map[int]string {
	0: "LOG_EMERG",
	1: "LOG_ALERT",
	2: "LOG_CRIT",
	3: "LOG_ERR",
	4: "LOG_WARNING",
	5: "LOG_NOTICE",
	6: "LOG_INFO",
	7: "LOG_DEBUG",
}

var severityMap = map[int]string {
	0:   "LOG_KERN",
	8:   "LOG_USER",
	16:  "LOG_MAIL",
	24:  "LOG_DAEMON",
	32:  "LOG_AUTH",
	40:  "LOG_SYSLOG",
	48:  "LOG_LPR",
	56:  "LOG_NEWS",
	64:  "LOG_UUCP",
	72:  "LOG_CRON",
	80:  "LOG_AUTHPRIV",
	88:  "LOG_FTP",
	128: "LOG_LOCAL0",
	136: "LOG_LOCAL1",
	144: "LOG_LOCAL2",
	152: "LOG_LOCAL3",
	160: "LOG_LOCAL4",
	168: "LOG_LOCAL5",
	176: "LOG_LOCAL6",
	184: "LOG_LOCAL7",
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
		Facility: prio / 8,
		Severity: prio % 8}
	m.Msg = string(bytes.TrimSpace(matches[2]))
	return m, nil
}
