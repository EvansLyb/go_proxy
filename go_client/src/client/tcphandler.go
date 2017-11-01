package client

import (
	"net"
	"log"
	"syscall"
	"strings"
	"client_crypt"
	"encoding/binary"
	"bytes"
)

func Tcpclient(port int, host string, rport int) error {
	var addr = net.TCPAddr{
		IP:   nil,
		Port: port,
		Zone: "",
	}

	listener, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Fatal(err)
	}

	fs, err := listener.File()
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	syscall.SetsockoptInt(int(fs.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	syscall.SetsockoptInt(int(fs.Fd()), syscall.SOL_IP, syscall.SO_KEEPALIVE, 1)

	for {
		if con, err := listener.AcceptTCP(); err != nil {

			Logger.Println("TCP accept error " + err.Error())

		} else {

			//handle ipv4
			if len([]int32(strings.Split(con.RemoteAddr().String(), ":")[0])) < 16 {

				go handle_tcp4_connecton(con, host, rport)
			}

			//handle ipv6
		}

	}

	return nil
}

func handle_tcp4_connecton(con *net.TCPConn, host string, rport int) {
	defer con.Close()

	dest_addr := con.LocalAddr()

	//convert dest addr to bytes
	dest, err := get_ipv4dest_bytes(dest_addr.String())

	if err != nil {
		Logger.Println("parse dest ip error " + err.Error())
		return
	}

	//connect to  the proxy server
	remote, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: rport,
		Zone: "",
	})
	if err != nil {
		Logger.Println("tcp : connect proxy server error from " + con.RemoteAddr().String() + ":" + err.Error())
		return
	}

	Logger.Println("tcp : from " + con.RemoteAddr().String() + " connectded to " + con.LocalAddr().String())

	defer remote.Close()

	//first write the 6 bytes to proxy server=====================================================================================

	//chacha20 encdata len 34 bytes
	if werr := client_crypt.Write_enc_data(remote, dest); werr != nil {
		Logger.Println("tcp : fail to write dest addr to proxy server from " + con.RemoteAddr().String() + ":" + err.Error())
		return
	}

	//==============================================================================================================================

	break_chan := make(chan bool)

	//send to remote recv from local
	send_chan := make(chan []byte, 100)

	//recv from remote send to local
	recv_chan := make(chan []byte, 100)

	//open a goroute to handle remote recv
	go func(remote *net.TCPConn, dest []byte, rchan chan<- []byte) {
		defer close(rchan)

		for {
			temp_buff := make([]byte, 0)
			//get enc data len
			data_len := make([]byte, 2)
			i, err := remote.Read(data_len)
			if err != nil || i != 2 {
				break_chan <- true
				return
			}
			//enc data length
			length := int(binary.BigEndian.Uint16(data_len))
			//totle recv data length
			recv_length := 0
			for {
				//read enc data
				enc_data := make([]byte, length-recv_length)
				i, rerr := remote.Read(enc_data)
				if rerr != nil {
					break_chan <- true
					return
				}
				recv := bytes.Join([][]byte{temp_buff, enc_data[:i]}, nil)

				if len(recv) == length {
					recv, err := client_crypt.Decrypt(recv[:i][12:], recv[:i][:12])

					if err != nil {
						break_chan <- true
						return
					}
					rchan <- recv
					break

				} else if len(recv) < length {
					recv_length += i
					continue
				} else {
					break_chan <- true
					return
				}
			}

		}
	}(remote, dest, recv_chan)

	//open to handle local recv
	go func(local *net.TCPConn, schan chan<- []byte) {
		defer close(schan)
		for {
			send := make([]byte, 20480)
			i, err := local.Read(send)
			if err != nil {
				break_chan <- true
				return
			}
			if i > 0 {

				schan <- send[:i]

			} else {
				break_chan <- true
				return
			}

		}
	}(con, send_chan)

	for {
		select {
		case recv_data := <-recv_chan:

			if _, err := con.Write(recv_data); err != nil {
				return

			}
		case send_data := <-send_chan:

			data_len := make([]byte, 2)
			dst, nonce := client_crypt.Encrypt(send_data)
			binary.BigEndian.PutUint16(data_len, uint16(len(dst)+len(nonce)))
			if _, err := remote.Write(bytes.Join([][]byte{data_len, nonce, dst}, nil)); err != nil {
				return
			}


		case <-break_chan:

			return

		}
	}

}
