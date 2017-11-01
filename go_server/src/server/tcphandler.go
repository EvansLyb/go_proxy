package server

import (
	"net"
	"log"
	"syscall"
	"strings"
	"encoding/binary"
	"strconv"
	"bytes"
	"server_crypt"

)

func Tcpserver(port int) {

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   nil,
		Port: port,
		Zone: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	file, _ := listener.File()

	syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.SO_KEEPALIVE, 1)
	for {
		con, err := listener.AcceptTCP()
		if err != nil {
			Logger.Println(err)

		} else {

			//handle ipv4
			if len([]int32(strings.Split(con.RemoteAddr().String(), ":")[0])) < 16 {

				go handle_tcp4_connecton(con)
			}
			//handle ipv6

		}
	}

}

func handle_tcp4_connecton(con *net.TCPConn) {
	defer con.Close()

	//read dest=========================================================================

	_, dest, err := server_crypt.Read_enc_data(con, 34)

	if err != nil {
		Logger.Println("can not read the dest addr : " + err.Error())
		return
	}

	//===============================================================

	dest_port := binary.BigEndian.Uint16(dest[:2])

	target, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.IPv4(dest[2], dest[3], dest[4], dest[5]),
		Port: int(dest_port),
		Zone: "",
	})

	Logger.Println("tcp : from " + con.RemoteAddr().String() + " connected to " + net.IPv4(dest[2], dest[3], dest[4], dest[5]).String() + ":" + strconv.Itoa(int(dest_port)))

	if err != nil {
		Logger.Println("can not connect to target addr from " + con.RemoteAddr().String() + " to " + net.IPv4(dest[2], dest[3], dest[4], dest[5]).String() + ":" + strconv.Itoa(int(dest_port)) + " : " + err.Error())
		return
	}
	defer target.Close()

	break_chan := make(chan bool)

	//send to target recv from client
	send_chan := make(chan []byte, 100)

	//recv from target send to client
	recv_chan := make(chan []byte, 100)

	go func(target *net.TCPConn, rchan chan<- []byte) {
		defer close(rchan)
		for {
			recv := make([]byte, 20480)
			i, err := target.Read(recv)
			if err != nil {
				break_chan <- true
				return
			}
			if i > 0 {
				rchan <- recv[:i]

			} else {
				break_chan <- true
				return
			}

		}
	}(target, recv_chan)

	go func(con *net.TCPConn, schan chan<- []byte) {
		defer close(schan)
		for {
			temp_buff := make([]byte, 0)
			//get enc data len
			data_len := make([]byte, 2)
			i, err := con.Read(data_len)
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
				i, rerr := con.Read(enc_data)
				if rerr != nil {
					break_chan <- true
					return
				}
				recv := bytes.Join([][]byte{temp_buff, enc_data[:i]}, nil)

				if len(recv) == length {
					recv, err := server_crypt.Decrypt(recv[:i][12:], recv[:i][:12])

					if err != nil {
						break_chan <- true
						return
					}
					schan <- recv
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
	}(con, send_chan)

	for {
		select {

		case recv_data := <-recv_chan:
			dst, nonce := server_crypt.Encrypt(recv_data)
			data_len := make([]byte, 2)
			binary.BigEndian.PutUint16(data_len, uint16(len(dst)+len(nonce)))
			if _,err:=con.Write(bytes.Join([][]byte{data_len, nonce, dst}, nil));err!=nil{
				return
			}

		case send_data := <-send_chan:

			if _, err := target.Write(send_data); err != nil {
				return
			}
		case <-break_chan:

			return

		}
	}

}
