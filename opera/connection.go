package opera

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net"
	"pos-proxy/config"
	"pos-proxy/db"
	"sync"
	"time"

	lock "github.com/bsm/redis-lock"
)

var conn net.Conn

const RETIES = 1

var mutex = &sync.Mutex{}

func doSendRequest(data []byte) (string, error) {
	if conn == nil {
		log.Println("Connection with opera died")
		return "", errors.New("Please check Opera POS Interface")
	}
	log.Println("About to lock")
	mutex.Lock()
	defer mutex.Unlock()
	log.Println("About to send message", string(data))
	stx := []byte{0x02}
	etx := []byte{0x03}
	payload := []byte{}
	payload = append(payload, stx...)
	payload = append(payload, data...)
	payload = append(payload, etx...)
	_, err := conn.Write(payload)
	if err != nil {
		log.Println("About to unlock in error")
		return "", err
	}
	log.Println("Will read buffer")
	message, err := bufio.NewReader(conn).ReadString(etx[0])
	log.Println("About to unlock in success")
	return message, err
}

func SendRequest(data []byte) (string, error) {
	if sendLinkStart() {
		return doSendRequest(data)
	} else {
		if conn != nil {
			conn.Close()
		}
		Connect()
		return doSendRequest(data)
	}
	return "", errors.New("Couldn't send link description")
}

func Connect() {
	var err error
	connectionString := fmt.Sprintf("%s:5010", config.Config.OperaIP)
	retries := 0
	connected := false
	log.Println("retries", retries, "connected", connected)
	for retries < RETIES && !connected {
		log.Println("retries", retries, "connected", connected)
		conn, err = net.Dial("tcp", connectionString)
		if err != nil {
			log.Println("Couldn't connect to opera with err ", err)
			log.Println("Connection string", connectionString)
		} else {
			connected = true
			break
		}
		retries += 1
		sleepValue := 1500
		time.Sleep(time.Duration(sleepValue) * time.Millisecond)
	}
	if !connected {
		log.Println("Couldn't connect to opera on ", connectionString)
		return
	}
	log.Println(fmt.Sprintf("Connection successful to Opera on %s", config.Config.OperaIP))
	sendLinkDescription()
}

func sendLinkStart() bool {
	t := time.Now()
	val := fmt.Sprintf("%02d%02d%02d", t.Year(), t.Month(), t.Day())
	val = val[2:]
	date := val

	val = fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second())
	time_value := val
	linkStartStr := fmt.Sprintf(`<LinkStart Date="%s" Time="%s" VerNum="1.0" />`, date, time_value)
	message, _ := doSendRequest([]byte(linkStartStr))
	if len(message) > 1 {
		message = message[1 : len(message)-1]
	}
	linkAlive := LinkAlive{}
	responseBuf := bytes.NewBufferString(message)
	if err := xml.NewDecoder(responseBuf).Decode(&linkAlive); err != nil {
		return false
	}
	return true
}

func sendLinkDescription() bool {
	t := time.Now()
	val := fmt.Sprintf("%02d%02d%02d", t.Year(), t.Month(), t.Day())
	val = val[2:]
	date := val

	val = fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second())
	time_value := val
	linkDescription := fmt.Sprintf(`<LinkDescription Date="%s" Time="%s" VerNum="1.0" />`, date, time_value)
	message, _ := doSendRequest([]byte(linkDescription))
	if len(message) > 1 {
		message = message[1 : len(message)-1]
	}
	linkAlive := LinkAlive{}
	responseBuf := bytes.NewBufferString(message)
	if err := xml.NewDecoder(responseBuf).Decode(&linkAlive); err != nil {
		return false
	}
	return true
}

func LockOpera() (*lock.Locker, error) {
	lockOptions := &lock.Options{
		WaitTimeout: 4 * time.Second,
	}

	l, err := lock.ObtainLock(db.Redis, "opera", lockOptions)
	if err != nil {
		return &lock.Locker{}, err
	} else if l == nil {
		return &lock.Locker{}, errors.New("couldn't obtain opera lock")
	}

	ok, err := l.Lock()
	if err != nil {
		return &lock.Locker{}, err
	} else if !ok {
		return &lock.Locker{}, errors.New("failed to acquire opera lock")
	}

	return l, nil
}
