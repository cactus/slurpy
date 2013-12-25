package main

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"github.com/cactus/gologit"
	"net"
	"os"
	"runtime"
	"strconv"
	"syscall"
)

const VERSION = "0.0.1"

func chanByteReader(ch <-chan *SyslogMsg) {
	for m := range ch {
		fmt.Printf("<%d> %s\n", m.Priority, m.Msg)
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
		BindTCP string `short:"t" long:"listen-tcp" default:"" description:"TCP address:port to listen to"`
		BindUDP string `short:"u" long:"listen-udp" default:"" description:"UDP address:port to listen to"`
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

	// set logger debug level and start toggle on signal handler
	logger := gologit.Logger
	logger.Set(opts.Verbose)
	logger.Debugln("Debug logging enabled")
	logger.ToggleOnSignal(syscall.SIGUSR1)

	ch := make(chan *SyslogMsg, 1024)
	go chanByteReader(ch)

	if opts.BindTCP != "" {
		gologit.Println("Starting tcp server on", opts.BindTCP)
		ltcp, err := net.Listen("tcp", opts.BindTCP)
		if err != nil {
			gologit.Fatal(err)
		}
		defer ltcp.Close()
		go tcpAcceptor(ltcp, ch)
	}

	if opts.BindUDP != "" {
		gologit.Println("Starting udp server on", opts.BindUDP)
		addr, err := net.ResolveUDPAddr("udp", opts.BindUDP)
		if err != nil {
			gologit.Fatal(err)
		}
		ludp, err := net.ListenUDP("udp", addr)
		if err != nil {
			gologit.Fatal(err)
		}
		defer ludp.Close()
		go udpAcceptor(ludp, ch)
	}
	select {}
}
