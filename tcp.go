package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	IPFX string = "[INFO] "
	EPFX string = "[ERROR] "
)

var (
	// argumentos de línea de comandos
	listen   bool
	quiet    bool
	addr     string
	proto    string
	port     uint
	buffsize uint
	errch    chan error = make(chan error, 1)
)

func main() {
	log.SetOutput(os.Stderr)

	if !parseArgs() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var listener net.Listener
	var conn net.Conn
	var err error

	if listen {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatalln(EPFX+"conexión fallida:", err)
		}
		log.Println(IPFX+"servidor iniciado en", addr)
		defer listener.Close()
	}

	conn, err = getConn(listener)
	if err != nil {
		log.Fatalln(EPFX+"conexión fallida:", err)
	}
	log.Printf("%sconectado a %s (%s)\n",
		IPFX,
		conn.RemoteAddr().String(),
		strings.Split(conn.LocalAddr().String(), ":")[1],
	)
	defer conn.Close()

	go readWriteLoop(conn, os.Stdout)
	go readWriteLoop(os.Stdin, conn)

	switch err = <-errch; err {
	case io.EOF:
		log.Println(IPFX+"conexión terminada:", err)
	default:
		log.Fatalln(EPFX+"error en la conexión:", err)
	}
}

func parseArgs() bool {
	var udp bool

	flag.BoolVar(&listen, "l", false, "Se queda a la escucha")
	flag.BoolVar(&quiet, "q", false, "No imprime mensajes de debug")
	flag.BoolVar(&udp, "u", false, "Utiliza UDP en lugar de TCP")
	flag.StringVar(&addr, "H", "", "Direccion IP")
	flag.UintVar(&buffsize, "b", 1, "Tamaño del búffer (en KB)")
	flag.UintVar(&port, "p", 8080, "Puerto")
	flag.Parse()

	if udp {
		proto = "udp"
	} else {
		proto = "tcp"
	}

	if quiet {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}

	addr = fmt.Sprintf("%s:%v", addr, port)

	return flag.Parsed()
}

func getConn(ln net.Listener) (net.Conn, error) {
	if ln != nil {
		return ln.Accept()
	} else {
		return net.Dial(proto, addr)
	}
}

func readWriteLoop(in io.Reader, out io.Writer) {
	buffer := make([]byte, 1024*buffsize)
	for {
		n, err := in.Read(buffer)
		if err != nil {
			errch <- err
			return
		}
		_, err = out.Write(buffer[:n])
		if err != nil {
			errch <- err
			return
		}
	}
}
