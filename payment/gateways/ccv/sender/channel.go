package sender

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	proxyEntity "pos-proxy/entity"
	"pos-proxy/payment/gateways/ccv/entity"
	"pos-proxy/payment/gateways/ccv/utils"
	"pos-proxy/socket"
)

// Channel holds the connection attributes
type Channel struct {
	host string
	port int
	conn *net.Conn
}

var channels = make(map[string]*Channel)

// Connect creates a new tcp connection with CCV pinpad
func Connect(settings proxyEntity.CCVSettings) (*Channel, error) {
	if _, ok := channels[settings.IP]; !ok {
		channels[settings.IP] = &Channel{}
	}
	channel := *channels[settings.IP]
	channel.host = settings.IP
	channel.port = settings.PinpadPort
	log.Println("starting a new connection with pinpad")
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", channel.host, channel.port))
	if err != nil {
		return &channel, err
	}
	channel.conn = &conn
	return &channel, err
}

// Send writes to the connection and wait for response
func Send(outputChan chan<- socket.Event, req *entity.SaleRequest, settings proxyEntity.CCVSettings) (*entity.SaleResponse, error) {
	c, err := getChannel(settings)
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
	payload, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	e := socket.Event{
		Module:  "payment",
		Type:    eventType,
		Payload: payload,
	}
	outputChan <- e
	// close the connection
	go closeConn(outputChan, conn)
	return &resp, nil
}

func getChannel(settings proxyEntity.CCVSettings) (*Channel, error) {
	if c, ok := channels[settings.IP]; !ok {
		if c.conn == nil {
			return Connect(settings)
		}
		return c, nil
	}

	channel, err := Connect(settings)
	return channel, err
}

func closeConn(outputChan chan<- socket.Event, conn *net.Conn) {
	log.Println("starting process of closing connection")
	err := utils.Send(conn, []byte{'F', 'I', 'N'})
	if err != nil {
		log.Println(err)
		return
	}
	// wait for response
	resp := make([]byte, 4)
	_, err = (*conn).Read(resp)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
			return
		}
	}
	log.Println("received:", resp)
	log.Println("sending ACK")
	err = utils.Send(conn, []byte{'A', 'C', 'K'})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("trying to close connection")
	(*conn).Close()
}
