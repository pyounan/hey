package main

import (
	"fmt"
	"log"
	"net"
	"pos-proxy/ccv/channel"
	"pos-proxy/ccv/entity"
	"time"
)

var fin = make(chan bool)

func main() {
	log.SetFlags(log.Lshortfile)
	var err error
	/*	commandChannel, err := channel.GetCommandChannel()
		if err != nil {
			log.Fatal(err)
		}
	*/
	deviceChannel, err := channel.GetDeviceChannel()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		req := entity.DeviceRequest{}
		req.ID = i
		channel.Send(deviceChannel, req)
		time.Sleep(1 * time.Second)
	}

	go handleFIN(fin)
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)
	ln, err := conn.Read(buff)
	if err != nil {
		log.Println(err)
		return
	}
	if fmt.Sprintf("%s", buff[:ln]) == "FIN" {
		fin <- true
	}
	conn.Write([]byte("ACK"))

	log.Printf("%s\n", buff)
}

/*
func handleFIN(fin chan bool) {
	for v := range fin {
		log.Println("RECEIVED FIN", v)
		commandChannel.FIN()
	}
}
*/

func handleC(c <-chan time.Time) {
	select {
	case v := <-c:
		log.Println(v)
		req := entity.DeviceRequest{}
		commandChannel.Send(req)
	}
}
