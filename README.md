slurpy
======

## About

A syslog slurper. Currently slurps udp or tcp.

[![Build Status](https://github.com/cactus/slurpy/workflows/unit-tests/badge.svg)](https://github.com/cactus/slurpy/actions)

It doesn't currently do anything other than parse the incoming syslog messages,
and print the result.

## Usage

Help output:

~~~
$ ./slurpy -h
Usage:
  slurpy [OPTIONS]

Application Options:
  -t, --listen-tcp= TCP address:port to listen to
  -u, --listen-udp= UDP address:port to listen to
  -v, --verbose     Show verbose (debug) log level output
  -V, --version     print version and exit

Help Options:
  -h, --help        Show this help message
~~~

Example running:

~~~
$ ./slurpy --listen-tcp="127.0.0.1:1514" --listen-udp="127.0.0.1:1514"
~~~

Example rsyslog config to forward to slurpy:

~~~
## tcp
*.* @@127.0.0.1:1514
## or udp
#*.* @127.0.0.1:1514
~~~

### Links

*   [RFC-3164](http://tools.ietf.org/html/rfc3164) - "The BSD syslog Protocol"
*   [RFC-6587](http://tools.ietf.org/html/rfc6587) - "Transmission of Syslog Messages over TCP"

## License

Released under an [ISC license][1]. See `LICENSE.md` file for details.

[1]: https://choosealicense.com/licenses/isc/
