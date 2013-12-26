package slurpylog

import (
	"testing"
)

var syslogMsgParseTests = []struct {
	SyslogText string
	Msg        SyslogMsg
}{
	{"<142>Dec 26 20:16:47 research1-west client-api[31801]: [INFO] mqueue.py:267 Connecting", SyslogMsg{Priority: 142, Facility: 136, Severity: 6, Msg: "[INFO] mqueue.py:267 Connecting"}},
	{"<142>Dec 26 20:16:47 research1-west client-api: [INFO] mqueue.py:267 Connecting", SyslogMsg{Priority: 142, Facility: 136, Severity: 6, Msg: "[INFO] mqueue.py:267 Connecting"}},
}

func TestMsgParsing(t *testing.T) {
	for _, tt := range syslogMsgParseTests {
		msg, err := parseSyslogMsg([]byte(tt.SyslogText))
		if err != nil {
			t.Fatal(err)
		}

		if msg.Priority != tt.Msg.Priority {
			t.Fatalf("Got priority '%d' expected '%d'", tt.Msg.Priority, tt.Msg.Priority)
		}

		if msg.Facility != tt.Msg.Facility {
			t.Fatalf("Got facility '%d' expected '%d'", tt.Msg.Facility, tt.Msg.Facility)
		}

		if msg.Severity != tt.Msg.Severity {
			t.Fatalf("Got severity '%d' expected '%d'", tt.Msg.Severity, tt.Msg.Severity)
		}
	}
}
