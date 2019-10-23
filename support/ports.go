package support

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type Port struct {
	Port         uint32 // What port number
	SSL          bool   // Does the port use SSL?
	FixedDiff    bool   // Is the port fixed diff
	StartingDiff uint32 // What's the starting difficulty
	MaxDiff      uint32 // If the port is fixed diff, this doesn't matter.
}

func (p *Port) Listen(c chan<- net.Conn) {
	var l net.Listener
	var err error
	if !p.SSL {
		l, err = net.Listen("tcp", fmt.Sprintf(":%v", p.Port))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		certHandle, err := os.Open("server.crt")
		if err != nil {
			log.Fatal(err)
		}
		certData, err := ioutil.ReadAll(certHandle)
		if err != nil {
			log.Fatal(err)
		}
		certHandle.Close()
		keyHandle, err := os.Open("server.key")
		if err != nil {
			log.Fatal(err)
		}
		keyData, err := ioutil.ReadAll(keyHandle)
		if err != nil {
			log.Fatal(err)
		}
		keyHandle.Close()
		cert, err := tls.X509KeyPair(certData, keyData)
		if err != nil {
			log.Fatal(err)
		}
		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		l, err = tls.Listen("tcp", fmt.Sprintf(":%v", p.Port), cfg)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			c <- nil
		}
		c <- conn
	}
}
