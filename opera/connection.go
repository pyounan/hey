package opera

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"pos-proxy/config"
	"time"
)

var conn net.Conn

func SendRequest(data []byte) {
	stx := []byte{0x02}
	etx := []byte{0x03}
	payload := []byte{}
	payload = append(payload, stx...)
	payload = append(payload, data...)
	payload = append(payload, etx...)
	conn.Write(payload)
	message, _ := bufio.NewReader(conn).ReadString(etx[0])
	fmt.Println("total size:", len(message), message)
}

func Connect() {
	var err error
	connectionString := fmt.Sprintf("%s:5010", config.Config.OperaIP)
	conn, err = net.Dial("tcp", connectionString)
	if err != nil {
		log.Println("Couldn't connect to opera with err ", err)
		return
	}
	log.Println(fmt.Sprintf("Connection successful to Opera on %s", config.Config.OperaIP))

	t := time.Now()
	val := fmt.Sprintf("%02d%02d%02d", t.Year(), t.Month(), t.Day())
	val = val[2:]
	date := val

	val = fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second())
	time_value := val
	linkDescription := fmt.Sprintf(`<LinkDescription Date="%s" Time="%s" VerNum="1.0" />`, date, time_value)
	SendRequest([]byte(linkDescription))
}
