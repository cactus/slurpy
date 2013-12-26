package syslog

type SyslogMsgHandler func (*SyslogMsg)

type SyslogServer interface {
	Close()
	Start()
}
