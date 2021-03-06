package receiver

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

type Channel struct {
	port int
	ln   *net.Listener
}

var channels = make(map[int]*Channel)
var output chan<- []string
var close chan bool
var incoming chan bool

func Listen(settings *proxyEntity.CCVSettings, notif chan<- socket.Event) error {
	channel, err := getChannel(*settings)
	if err != nil {
		return err
	}
	go handleClosing()
	go handleIncomingMessages(channel.ln, notif)
	return nil
}

func getChannel(settings proxyEntity.CCVSettings) (*Channel, error) {
	if _, ok := channels[settings.ProxyPort]; !ok {
		return newChannel(settings.ProxyPort)
	}
	channel := channels[settings.ProxyPort]
	if channel.ln == nil {
		return newChannel(settings.ProxyPort)
	}
	return channel, nil
}

func newChannel(port int) (*Channel, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	if _, ok := channels[port]; !ok {
		channels[port] = &Channel{
			port: port,
		}
	}
	channels[port].ln = &ln
	return channels[port], nil
}

func handleIncomingMessages(c *net.Listener, notif chan<- socket.Event) {
	for {
		select {
		case _ = <-close:
			break
		default:
			conn, err := (*c).Accept()
			if err != nil {
				continue
			}
			processMessage(&conn, notif)
		}
	}
}

func handleInput(conn *net.Conn, resp entity.DeviceRequest) error {
	req := entity.DeviceResponse{}
	req.Attrs.RequestType = "Input"
	req.Attrs.RequestID = resp.Attrs.RequestID
	req.Attrs.WorkstationID = resp.Attrs.WorkstationID
	req.Attrs.OverallResult = "Success"
	req.Output.Target = resp.Output.Target
	req.Output.OutResult = "Success"
	req.Attrs.XMLNS = resp.Attrs.XMLNS
	req.Input.Target = "CashierKeyboard"
	req.Input.InResult = "Success"
	req.Input.InputValue.InNumber = 1
	return sendToChan(conn, &req)
}

func processMessage(conn *net.Conn, notif chan<- socket.Event) (*entity.DeviceRequest, error) {
	defer (*conn).Close()
	resp := entity.DeviceRequest{}

	msgLen := make([]byte, 4)
	_, err := (*conn).Read(msgLen)
	if err != nil {
		return &resp, err
	}
	n := binary.BigEndian.Uint32(msgLen)
	result := []byte{}
	ln := 0
	for ln < int(n) {
		tmp := make([]byte, 1024)
		nRead, err := (*conn).Read(tmp)
		if err != nil {
			if err != io.EOF {
				return &resp, err
			}
			break
		}
		ln += nRead
		result = append(result, tmp[:nRead]...)
	}
	log.Printf("received : %s", result)
	// encode payload to xml of type DeviceResponse
	buff := bytes.NewBuffer(result)
	err = xml.NewDecoder(buff).Decode(&resp)
	if err != nil {
		log.Println(err)
		return &resp, err
	}
	m := socket.Event{}
	m.Module = "payment"
	m.Type = "output"
	// var payload interface{}
	type outputPayload struct {
		Target string      `json:"target"`
		Body   interface{} `json:"body"`
	}
	payload := outputPayload{}
	if resp.Output.Target == "CashierDisplay" || resp.Output.Target == "Printer" {
		payload.Body = []string{}
		payload.Target = resp.Output.Target
		for _, t := range resp.Output.TextLines {
			payload.Body = append(payload.Body.([]string), t.Text)
		}
	} else if resp.Output.Target == "E-Journal" && resp.Output.EJournal != nil {
		payload.Body = *resp.Output.EJournal
		payload.Target = resp.Output.Target
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		p := make(map[string]string, 1)
		p["error"] = err.Error()
		encodedPayload, _ := json.Marshal(p)
		e := socket.Event{
			Module:  "payment",
			Type:    "error",
			Payload: encodedPayload,
		}
		notif <- e
		return nil, err
	}
	m.Payload = encodedPayload
	notif <- m
	log.Println("=======================")
	// time.Sleep(1 * time.Second)

	if resp.RequestType == "Input" {
		err = handleInput(conn, resp)
		return &resp, err
	}

	req := entity.DeviceResponse{}
	req.Attrs.RequestType = "Output"
	req.Attrs.RequestID = resp.Attrs.RequestID
	req.Attrs.WorkstationID = resp.Attrs.WorkstationID
	req.Attrs.OverallResult = "Success"
	req.Output.Target = resp.Output.Target
	req.Output.OutResult = "Success"
	req.Attrs.XMLNS = resp.Attrs.XMLNS
	log.Println("prepairing response to DeviceRequest")
	err = sendToChan(conn, &req)
	if err != nil {
		log.Println(err)
	}
	return &resp, nil
}

func handleClosing() {
	for range close {
		log.Println("close")
	}
}

// Send writes to the connection and wait for response
func Send(resp *entity.DeviceResponse, settings proxyEntity.CCVSettings) error {
	c, err := getChannel(settings)
	if err != nil {
		return err
	}
	// decode the struct to xml
	buff := bytes.NewBuffer([]byte{})
	err = xml.NewEncoder(buff).Encode(resp)
	if err != nil {
		log.Println(err)
		return err
	}
	conn, err := (*c.ln).Accept()
	if err != nil {
		log.Println("error sending my response to device request")
		log.Println(err)
	}
	log.Printf("sending %s\n", buff)
	err = utils.Send(&conn, buff.Bytes())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func sendToChan(conn *net.Conn, resp *entity.DeviceResponse) error {
	// decode the struct to xml
	buff := bytes.NewBuffer([]byte{})
	err := xml.NewEncoder(buff).Encode(resp)
	if err != nil {
		log.Println(err)
		return err
	}

	// log.Printf("sending %s\n", buff)
	err = utils.Send(conn, buff.Bytes())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
