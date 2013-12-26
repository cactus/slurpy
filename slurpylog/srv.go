package slurpylog

type SyslogMsgHandler func(*SyslogMsg)

type SyslogServer interface {
	Close()
	Start()
}
