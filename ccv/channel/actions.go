package channel

import (
	"log"
	"net"
)

func Send(channel *net.Conn, payload []byte) error {
	log.Printf("should send %s\n", payload)
	return nil
}
