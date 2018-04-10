package sender

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"pos-proxy/payment/gateways/ccv/entity"
	"pos-proxy/payment/gateways/ccv/utils"
	"pos-proxy/socket"
)

type Channel struct {
	host string
	port string
	conn *net.Conn
}

var channel = &Channel{}

func Connect(host, port string) (*Channel, error) {
	channel.host = host
	channel.port = port
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", channel.host, channel.port))
	if err != nil {
		return channel, err
	}
	channel.conn = &conn
	return channel, err
}

// Send writes to the connection and wait for response
func Send(outputChan chan<- socket.Event, req *entity.SaleRequest) (*entity.SaleResponse, error) {
	c, err := getChannel()
	if err != nil {
		return &entity.SaleResponse{}, err
	}
	// decode the struct to xml
	buff := bytes.NewBuffer([]byte{})
	err = xml.NewEncoder(buff).Encode(req)
	if err != nil {
		return &entity.SaleResponse{}, err
	}
	err = utils.Send(c.conn, buff.Bytes())
	if err != nil {
		return &entity.SaleResponse{}, err
	}
	// wait for response
	resp, err := read(outputChan, c.conn)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func read(outputChan chan<- socket.Event, conn *net.Conn) (*entity.SaleResponse, error) {
	resp := entity.SaleResponse{}

	msgLen := make([]byte, 4)
	_, err := (*conn).Read(msgLen)
	if err != nil {
		return &resp, err
	}
	n := binary.BigEndian.Uint32(msgLen)
	log.Printf("receiving message of length: %d", n)
	result := make([]byte, n)
	_, err = (*conn).Read(result)
	if err != nil {
		return &resp, err
	}
	log.Printf("%s\n", result)
	// encode payload to xml of type SaleResponse
	buff := bytes.NewBuffer(result)
	err = xml.NewDecoder(buff).Decode(&resp)
	if err != nil {
		return &resp, err
	}
	eventType := "error"
	if resp.Attrs.OverallResult == "Success" {
		eventType = "success"
	} else {
		eventType = resp.Attrs.OverallResult
	}
	e := socket.Event{
		Module: "payment",
		Type:   eventType,
	}
	outputChan <- e
	// close the connection
	go close(outputChan)
	return &resp, nil
}

func getChannel() (*Channel, error) {
	var err error
	if channel.conn == nil {
		channel, err = Connect(channel.host, channel.port)
	}
	return channel, err
}

func close(outputChan chan<- socket.Event) {
	log.Println("trying to close connection")
	ch, _ := getChannel()
	err := utils.Send(ch.conn, []byte{'F', 'I', 'N'})
	if err != nil {
		log.Println(err)
		return
	}
	// wait for response
	resp := make([]byte, 4)
	_, err = (*ch.conn).Read(resp)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
			return
		}
	}
	log.Println("closing, received:", resp)
	err = utils.Send(ch.conn, []byte{'A', 'C', 'K'})
	if err != nil {
		log.Println(err)
		return
	}
	(*ch.conn).Close()
}
