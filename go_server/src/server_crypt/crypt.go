package server_crypt

import (
	"crypto/cipher"
	"math/rand"
	"time"
	"encoding/binary"
	"bytes"
	"io"
	"errors"
	"fmt"
)

var (
	Aead cipher.AEAD
)

func Encrypt(data []byte) (dst []byte, nonce []byte) {
	nonce = make([]byte, 12)
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)
	binary.BigEndian.PutUint64(nonce, rnd.Uint64())
	dst = Aead.Seal(nil, nonce, data, []byte("fuckgfw"))
	return
}

func Decrypt(data, nonce []byte) (decdata []byte, err error) {

	decdata, err = Aead.Open(nil, nonce, data, []byte("fuckgfw"))
	return
}

func Write_enc_data(con io.Writer, data []byte) error {
	dst, nonce := Encrypt(data)
	_, err := con.Write(bytes.Join([][]byte{nonce, dst}, nil))
	fmt.Println(len(dst)+len(nonce))
	return err
}

func Read_enc_data(con io.Reader, buff int) (i int, data []byte, err error) {
	recv := make([]byte, buff)

	i, err = con.Read(recv)
	if err != nil {

		return i, nil, err
	}

	if i <= 28 {
		return -1, nil, errors.New("null data")
	}

	data, err = Decrypt(recv[:i][12:], recv[:i][:12])

	return

}
