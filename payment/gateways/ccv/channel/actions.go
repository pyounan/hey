package channel

import (
	"bytes"
	"cloudinn/proxy_plugins/ccv/entity"
	"encoding/binary"
	"encoding/xml"
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
	//payload = append([]byte(`<?xml version="1.0" encoding="utf-8"?>`), payload...)
	msgLen := toLengthIndicator(len(payload))
	payload = append(msgLen, payload...)
	log.Printf("should send %s \n to conn %v\n", payload, *conn)
	(*conn).Write(payload)
	/*msg := make([]byte, 1024)
	_, err := (*conn).Read(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("%s\n", msg)*/
	return nil
}

func SendAndWaitForResponse(conn *net.Conn, payload []byte) ([]byte, error) {
	msgLen := toLengthIndicator(len(payload))
	payload = append(msgLen, payload...)
	log.Printf("should send %s \n to conn %v\n", payload, *conn)
	(*conn).Write(payload)
	resp := Receive(conn)
	log.Printf("%s\n", resp)
	return resp, nil
}

func ACK(conn *net.Conn) {
	Send(conn, []byte("ACK"))
}

func FIN(conn *net.Conn) {
	Send(conn, []byte("FIN"))
}

func Receive(conn *net.Conn) []byte {
	msgLen := make([]byte, 4)
	_, err := (*conn).Read(msgLen)
	if err != nil {
		log.Println(err)
	}
	n := binary.BigEndian.Uint32(msgLen)
	log.Printf("receiving message of length: %d", n)
	result := make([]byte, n)
	_, err = (*conn).Read(result)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%s", result)
	resp, err := handleDeviceRequest(result)
	if err != nil {
		log.Println(err)
	}
	Send(conn, resp)
	return result
}

func handleDeviceRequest(req []byte) ([]byte, error) {
	data := entity.DeviceRequest{}
	reader := bytes.NewReader(req)
	err := xml.NewDecoder(reader).Decode(&data)
	if err != nil {
		log.Println(err)
		return []byte{}, err
	}
	log.Printf("#%v\n", data)
	for _, text := range data.Output.TextLines {
		log.Println(text.Text)
	}
	resp := entity.DeviceResponse{}
	resp.RequestID = data.RequestID
	resp.WorkstationID = data.WorkstationID
	resp.RequestType = data.RequestType
	resp.OverallResult = "Success"
	resp.XMLNS = data.XMLNS
	resp.Output.Target = data.Output.Target
	resp.Output.OutResult = "Success"
	b := bytes.NewBuffer([]byte{})
	err = xml.NewEncoder(b).Encode(resp)
	if err != nil {
		log.Println(err)
		return b.Bytes(), err
	}
	return b.Bytes(), nil
}
