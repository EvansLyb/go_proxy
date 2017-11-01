package client

import (
	"encoding/binary"
	"strings"
	"strconv"
	"net"
)

func get_ipv4dest_bytes(ip string) ([]byte, error) {
	dest := make([]byte, 2)

	ipa := strings.Split(ip, ":")
	dport, err := strconv.ParseInt(ipa[1], 10, 64)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint16(dest, uint16(dport))

	dip := net.ParseIP(ipa[0]).To4()

	return append(dest, dip...), nil
}
