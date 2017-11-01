package client

import (
	"net"
	"encoding/binary"
)

//udp header include the pseudo header

type ipv4_udp_header struct {
	source_ip net.IP
	dest_ip net.IP
	proto uint16
	totle_len uint16

	//follow is the data in fact to send

	source_port uint16
	dest_port uint16
	totle_length uint16
	checksum uint16

}


func (header *ipv4_udp_header)calculate_checksum(){

}

func (header *ipv4_udp_header)get_send_byte() []byte{
	data:=make([]byte,8)
	binary.BigEndian.PutUint16(data[:2],header.source_port)
	binary.BigEndian.PutUint16(data[2:4],header.dest_port)
	binary.BigEndian.PutUint16(data[4:6],header.totle_length)
	binary.BigEndian.PutUint16(data[6:],header.checksum)
	return data
}

