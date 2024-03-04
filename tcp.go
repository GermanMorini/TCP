package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"
)

const (
	iPfx string = "[INFO] "
	ePfx string = "[ERROR] "
)

var (
	listen bool
	quiet  bool
	addr   string
	ok     chan struct{} = make(chan struct{}, 1)
)

func parseArgs() bool {
	flag.BoolVar(&listen, "l", false, "Se queda a la escucha")
	flag.BoolVar(&quiet, "q", false, "No imprime mensajes de debug")
	flag.StringVar(&addr, "H", ":8080", "Direccion de la forma <IP>:<puerto>")

	flag.Parse()

	if quiet {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}

	return flag.Parsed()
}

func main() {
	if !parseArgs() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conn, listener, err := getConn()
	if err != nil {
		log.Println(ePfx+"conexión fallida:", err)
		os.Exit(2)
	}
	log.Println(iPfx+"conectado a", conn.RemoteAddr())

	go handle(conn, os.Stdin)
	go handle(os.Stdout, conn)

	<-ok
	close(conn, listener)
}

func close(c ...io.Closer) {
	for _, i := range c {
		if i != nil {
			i.Close()
		}
	}
}

func getConn() (net.Conn, net.Listener, error) {
	if listen {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, nil, err
		}
		log.Println(iPfx+"servidor iniciado en", addr)

		conn, err := listener.Accept()
		return conn, listener, err
	} else {
		conn, err := net.Dial("tcp", addr)
		return conn, nil, err
	}
}

func handle(out io.Writer, in io.Reader) {
	reader := bufio.NewReader(in)
	writer := bufio.NewWriter(out)

	for {
		data, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				log.Println(iPfx+"conexión terminada:", err)
				ok <- struct{}{}
				return
			}
			log.Println(ePfx+"conexión terminada:", err)
			ok <- struct{}{}
			return
		}

		writer.WriteString(data)
		writer.Flush()
	}
}
