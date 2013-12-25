package main

import (
	//"net"
	"strconv"
	flags "github.com/jessevdk/go-flags"
	"net"
	"log"
	"bufio"
	"bytes"
	"strings"
	"os"
	"fmt"
	"regexp"
	"runtime"
	"errors"
)


const VERSION = "0.0.1"
var matcher = regexp.MustCompile(`^(<(\d+)>)?(.*)`)

type SyslogMsg struct {
	Priority int
	Facility int
	Severity int
	Msg		 string
}

func chanByteReader(ch <-chan *SyslogMsg) {
	for m := range ch {
		fmt.Printf("<%d> %s\n", m.Priority, m.Msg)
	}
}


func parseSyslogMsg(buf []byte) (*SyslogMsg, error) {
	matches := matcher.FindSubmatch(buf)
	if len(matches) == 0 {
		return nil, errors.New("No match")
	}

	var prio int
	var err error
	if len(matches[2]) != 0 {
		prio, err = strconv.Atoi(string(matches[2]))
		if err != nil {
			return nil, errors.New("prio failed to convert")
		}
	} else {
		// default prio a relay must write if none is readable
		prio = 13
	}

	m := &SyslogMsg{
		Priority: prio,
		Facility: prio/8,
		Severity: prio % 8}
	m.Msg = string(bytes.Trim(matches[3], "\n"))
	return m, nil
}


func udpAcceptor(pc *net.UDPConn, ch chan<- *SyslogMsg) {
	buf := make([]byte, 1024)
	for {
		n, _, err := pc.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf("read %d\n", n)
		if n == 0 {
			continue
		}
		m, err := parseSyslogMsg(buf[0:n])
		if err != nil {
			continue
		}
		ch <- m
	}
}


func handleConn(c net.Conn, ch chan<- *SyslogMsg) {
	bufc := bufio.NewReader(c)
	defer c.Close()

	pk, err := bufc.Peek(1)
	if err != nil {
		return
	}
	if !bytes.Equal(pk, []byte("<")) {
		for {
			s, err := bufc.ReadString(' ')
			if err != nil {
				log.Println(err)
				break
			}
			s = strings.Trim(s, " ")
			readsize, err := strconv.Atoi(s)
			if err != nil {
				log.Println(err)
				break
			}

			var result []byte
			for i := 0; i < readsize; i++ {
				b, err := bufc.ReadByte()
				if err != nil {
					log.Println(err)
					break
				}
				result = append(result, b)
			}
			if len(result) > 0 {
				m, err := parseSyslogMsg(result)
				if err != nil {
					continue
				}
				ch <- m
			}
		}
	} else {
		for {
			result, err := bufc.ReadBytes('\n')
			if err != nil {
				log.Println(err)
				break
			}
			result = bytes.Trim(result, " ")
			if len(result) > 0 {
				m, err := parseSyslogMsg(result)
				if err != nil {
					log.Println(err)
					continue
				}
				ch <- m
			}
		}
	}
}

func tcpAcceptor(ln net.Listener, ch chan<- *SyslogMsg) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("got conn")
		go handleConn(conn, ch)
	}
}


func main() {
	var gmx int
	if gmxEnv := os.Getenv("GOMAXPROCS"); gmxEnv != "" {
		gmx, _ = strconv.Atoi(gmxEnv)
	} else {
		gmx = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(gmx)

	// command line flags
	var opts struct {
		BindTCP string `long:"listen-tcp" default:"" description:"TCP address:port to listen to"`
		BindUDP string `long:"listen-udp" default:"" description:"UDP address:port to listen to"`
		Verbose bool   `short:"v" long:"verbose" description:"Show verbose (debug) log level output"`
		Version bool   `short:"V" long:"version" description:"print version and exit"`
	}

	// parse said flags
	_, err := flags.Parse(&opts)
	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("slurpy-%s (%s,%s-%s)\n", VERSION, runtime.Version(), runtime.Compiler, runtime.GOARCH)
		os.Exit(0)
	}

	if opts.BindTCP == "" && opts.BindUDP == "" {
		fmt.Println("No listen ports defined. Exiting.")
		os.Exit(1)
	}

	ch := make(chan *SyslogMsg, 1024)
	go chanByteReader(ch)

	if opts.BindTCP != "" {
		ltcp, err := net.Listen("tcp", opts.BindTCP)
		if err != nil {
			log.Fatal(err)
		}
		defer ltcp.Close()
		go tcpAcceptor(ltcp, ch)
	}

	if opts.BindUDP != "" {
		addr, err := net.ResolveUDPAddr("udp", opts.BindUDP)
		if err != nil {
			log.Fatal(err)
		}
		ludp, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer ludp.Close()
		go udpAcceptor(ludp, ch)
	}
	select {}
}
