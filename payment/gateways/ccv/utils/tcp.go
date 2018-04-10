package utils

import (
	"encoding/binary"
	"log"
	"net"
)

// toLengthIndicator convers a message length from int to a byte array of length 4
func toLengthIndicator(length int) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(length))
	return bs
}

func Send(conn *net.Conn, payload []byte) error {
	msgLen := toLengthIndicator(len(payload))
	payload = append(msgLen, payload...)
	log.Printf("should send %s \n to conn %v\n", payload, *conn)
	_, err := (*conn).Write(payload)
	return err
}
