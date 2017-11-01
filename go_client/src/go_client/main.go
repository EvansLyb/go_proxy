package main

/*
this software use to bypass the gfw in chain.
it make use of linux iptables TPROXY and ip rule ,ip route to redirect any connection to local and encrypt the data send to remote server.
if have any question and want to communicate about the network programming,welcome mail to :906907952@qq.com .
ps:English is not my mother language.
 */

import (
	"flag"
	"fmt"
	"strconv"
	"time"
	"crypt"
	"bytes"
	"golang.org/x/crypto/chacha20poly1305"
	"log"
	"os"
	"client"
)

var (
	port = 0
	//remote proxy server listening port
	rport    = 0
	host     string
	password string
	help     bool
)

func init() {

	//parse command line param
	flag.BoolVar(&help, "help", false, "help manul")
	flag.StringVar(&password, "passwd", "", "decrypt/encrypt Password")
	flag.IntVar(&port, "port", 0, "listening port")
	flag.IntVar(&rport, "rport", 0, "remote host port")
	flag.StringVar(&host, "host", "", "remote host ip")
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if len([]int32(password)) <= 32 && len([]int32(password)) > 0 {
		fill := make([]byte, 32-len(password))
		pass := bytes.Join([][]byte{[]byte(password), fill}, nil)

		cipaead, err := chacha20poly1305.New(pass)
		if err != nil {
			log.Fatal(err)
		}
		crypt.Aead = cipaead

	} else {
		fmt.Println("use --help to useage")
		log.Fatal("Password length must be 0-32 bytes")
	}

}

func main() {

	fmt.Println("start client listen on 0.0.0.0:" + strconv.Itoa(port) + ",remote server " + host + ":" + strconv.Itoa(rport))
	client.Logger.Println("start client listen on 0.0.0.0:" + strconv.Itoa(port) + ",remote server " + host + ":" + strconv.Itoa(rport))

	go client.Tcpclient(port, host, rport)

	go client.Udpclient(port, host, rport)

	for {
		time.Sleep(999999999)
	}

}
