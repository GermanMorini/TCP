package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var (
	errch chan error = make(chan error, 1)

	// argumentos de línea de comandos
	listen   bool
	quiet    bool
	enc      bool
	addr     string
	proto    string
	port     uint
	buffsize uint
)

const (
	DEFAULT_PORT      uint = 8080
	DEFAULT_BUFF_SIZE uint = 1
)

// implementación de net.Listener para una conexión UDP
type UDPListener struct{}

func (l *UDPListener) Accept() (net.Conn, error) {
	conn, err := net.ListenPacket("udp", addr)
	return &UDPConn{conn: conn}, err
}

func (l *UDPListener) Close() error   { return nil }
func (l *UDPListener) Addr() net.Addr { return nil }

// implementación de net.Conn para una conexión UDP
type UDPConn struct {
	conn net.PacketConn
	addr net.Addr
}

func (c *UDPConn) Read(p []byte) (int, error) {
	n, addr, err := c.conn.ReadFrom(p)
	c.addr = addr
	return n, err
}

func (c *UDPConn) Write(p []byte) (int, error) {
	return c.conn.WriteTo(p, c.addr)
}

func (c *UDPConn) Close() error                       { return c.conn.Close() }
func (c *UDPConn) LocalAddr() net.Addr                { return c.conn.LocalAddr() }
func (c *UDPConn) RemoteAddr() net.Addr               { return c.addr }
func (c *UDPConn) SetDeadline(t time.Time) error      { return nil }
func (c *UDPConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *UDPConn) SetWriteDeadline(t time.Time) error { return nil }

func main() {
	if !parseArgs() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	var listener net.Listener
	var conn net.Conn

	if !enc {
		log.Println("Advertencia: la conexión NO es segura. Activá el cifrado con '-e'")
	}
	log.Println("PID del proceso:", os.Getpid())
	if listen {
		listener, err = getListener()
		if err != nil {
			log.Fatalln("Error iniciando servidor:", err)
		}
		log.Println("Servidor escuchando en", addr)
		defer listener.Close()
	}

	conn, err = getConn(listener)
	if err != nil {
		log.Fatalln("Error conectando:", err)
	}
	if proto != "udp" {
		log.Printf("Conexión establecida con %s (%s)\n",
			conn.RemoteAddr().String(),
			strings.Split(conn.LocalAddr().String(), ":")[1],
		)
	}
	defer conn.Close()

	go readWriteLoop(conn, os.Stdout)
	go readWriteLoop(os.Stdin, conn)

	switch err = <-errch; err {
	case io.EOF:
		log.Println("Conexión cerrada:", err)
	default:
		log.Fatalln("Error en la conexión:", err)
	}
}

func parseArgs() bool {
	var udp bool

	flag.BoolVar(&listen, "l", false, "Modo servidor")
	flag.BoolVar(&quiet, "q", false, "Silenciar logs")
	flag.BoolVar(&udp, "u", false, "Usar UDP")
	flag.BoolVar(&enc, "e", false, "Encriptar conexión")
	flag.StringVar(&addr, "H", "", "Dirección IP")
	flag.UintVar(&buffsize, "b", DEFAULT_BUFF_SIZE, "Tamaño de búfer (KB)")
	flag.UintVar(&port, "p", DEFAULT_PORT, "Puerto")
	flag.Parse()

	proto = "tcp"
	if udp {
		proto = "udp"
	}

	if quiet {
		log.SetOutput(io.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	addr = fmt.Sprintf("%s:%v", addr, port)
	return flag.Parsed()
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generando clave RSA: %v", err)
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<60))
	if err != nil {
		return nil, fmt.Errorf("error generando número de serie: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Auto-generated cert"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&key.PublicKey,
		key,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando certificado: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{certDER},
			PrivateKey:  key,
		}},
	}, nil
}

func getListener() (net.Listener, error) {
	if proto == "udp" {
		return &UDPListener{}, nil
	}

	if enc {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		tlsConfig, err := generateTLSConfig()
		if err != nil {
			return nil, err
		}

		return tls.NewListener(ln, tlsConfig), nil
	}

	return net.Listen("tcp", addr)
}

func getConn(ln net.Listener) (net.Conn, error) {
	if ln != nil {
		return ln.Accept()
	}

	if enc {
		return tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: true,
		})
	}
	return net.Dial(proto, addr)
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
