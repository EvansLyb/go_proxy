package server

import (
	"net"

	"strconv"
	"strings"
	"fmt"
	"bytes"
	"server_crypt"

)

func Udp_server(port int) {

	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: port,
		Zone: "",
	})

	if err != nil {
		Logger.Println("can not listen in port :" + strconv.Itoa(port) + ":" + err.Error())
	}
	defer listen.Close()




	for {

		send_data := make([]byte, 1430)
		i, addr, err := listen.ReadFromUDP(send_data)

		if err != nil {
			continue
		}

		if len([]int32(strings.Split(addr.String(), ":")[0])) < 16 {
			go handle_ipv4_udp_data(addr, send_data[:i], listen)
		}
	}
}

func handle_ipv4_udp_data(udp_addr *net.UDPAddr, data []byte, server *net.UDPConn) {

	data, err := server_crypt.Decrypt(data[12:], data[:12])

	if err != nil {
		fmt.Println("can not dec data from " + udp_addr.String() + " : " + err.Error())
		Logger.Println("can not dec data from " + udp_addr.String() + " : " + err.Error())
		return
	}

	dest_port := data[:2]
	port, _ := strconv.ParseInt(strconv.Itoa(int(dest_port[0]))+strconv.Itoa(int(dest_port[1])), 10, 16)

	remote, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(data[2], data[3], data[4], data[5]),
		Port: int(port),
		Zone: "",
	})

	if err != nil {
		Logger.Println("udp : can not connect to " + net.IPv4(data[2], data[3], data[4], data[5]).String() + ":" + err.Error())
		return
	}

	defer remote.Close()
	if _, werr := remote.Write(data[6:]); werr != nil {
		return
	}

	Logger.Println("udp : from " + udp_addr.IP.String() + " connected to " + remote.RemoteAddr().String())

	recv := make([]byte, 1464)
	i, rerr := remote.Read(recv)
	if rerr != nil {
		return
	}

	if i > 0 {
		dst,nonce:= server_crypt.Encrypt(recv[:i])
		fmt.Println(len(dst)+len(nonce))

		if _, werr := server.WriteToUDP(bytes.Join([][]byte{nonce,dst},nil), udp_addr); werr != nil {
			return
		}
	}

}
