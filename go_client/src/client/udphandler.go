package client

import (
	"net"
	"log"
	"syscall"
	"strings"
	"bytes"
	"golang.org/x/net/ipv4"
	"encoding/binary"

	"client_crypt"
)

func Udpclient(port int, host string, rport int) {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: port,
		Zone: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	defer listen.Close()

	file, err := listen.File()
	if err != nil {
		log.Fatal(err)
	}
	syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1)
	syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.SO_REUSEADDR, 1)

	for {
		data := make([]byte, 1430)
		//store the additional data include the dest addr
		oob := make([]byte, 1024)

		i, oobi, _, addr, err := listen.ReadMsgUDP(data, oob)

		if err != nil {
			continue
		}

		if len([]int32(strings.Split(addr.String(), ":")[0])) < 16 {

			go handle_ipv4_udp_data(listen, addr, data[:i], oob[:oobi][18:24], host, rport)
		}

	}

}

func handle_ipv4_udp_data(local *net.UDPConn, udp_addr *net.UDPAddr, data, oob []byte, remote_host string, remote_port int) {
	dest_byte := oob

	remote, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(remote_host),
		Port: remote_port,
		Zone: "",
	})

	if err != nil {
		Logger.Println("udp : connect proxy server error from " + udp_addr.String() + ":" + err.Error())
		return
	}

	defer remote.Close()

	Logger.Println("udp : from " + udp_addr.String() + " connected to " + remote.RemoteAddr().String())

	////send enc data===============================================================================
	//if _, werr := remote.Write(bytes.Join([][]byte{dest_byte, data}, nil)); werr != nil {
	//	return
	//}
	werr := client_crypt.Write_enc_data(remote, bytes.Join([][]byte{dest_byte, data}, nil))
	if werr != nil {
		return
	}

	//read data and dec============================================================================================

	//recv := make([]byte, 1464)
	//i, err := remote.Read(recv)
	//if err != nil {
	//	return
	//}

	_, recv, rerr := client_crypt.Read_enc_data(remote, 1464)

	if rerr != nil {
		return
	}

	i := len(recv)

	//construct the udp_header and ip header to send
	if i > 0 {

		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_UDP)
		if err != nil {
			return
		}
		if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_HDRINCL, 1); err != nil {
			return
		}

		defer syscall.Close(fd)

		src_ip := udp_addr.IP.To4()
		//construct the ip header
		ip_header := ipv4.Header{
			Version:  ipv4.Version,
			Len:      ipv4.HeaderLen,
			TOS:      0,
			TotalLen: ipv4.HeaderLen + i,
			ID:       0,
			Flags:    0,
			FragOff:  0,
			TTL:      64,
			Protocol: 17,
			Checksum: 0,
			Src:      net.IPv4(dest_byte[2], dest_byte[3], dest_byte[4], dest_byte[5]),
			Dst:      net.IPv4(src_ip[0], src_ip[1], src_ip[2], src_ip[3]),
			Options:  nil,
		}

		//construct the udp header
		udp_header := &ipv4_udp_header{

			source_ip:    dest_byte[2:],
			dest_ip:      udp_addr.IP.To4(),
			proto:        17,
			totle_len:    uint16(i + 8),
			source_port:  binary.BigEndian.Uint16(dest_byte[:2]),
			dest_port:    uint16(udp_addr.Port),
			totle_length: uint16(i + 8),
			checksum:     0,
		}

		udp_header.get_send_byte()
		ip_header_bytes, err := ip_header.Marshal()
		if err != nil {
			return
		}

		if serr := syscall.Sendto(

			fd, bytes.Join([][]byte{

				ip_header_bytes, //ip_header

				udp_header.get_send_byte(), //udp_header

				recv} /*data*/ , nil), 0,

			&syscall.SockaddrInet4{
				Port: udp_addr.Port,
				Addr: [4]byte{src_ip[0], src_ip[1], src_ip[2], src_ip[3]},
			}); serr != nil {

			return
		}

	}

}
