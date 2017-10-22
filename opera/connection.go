package opera

import (
	"bufio"
	"errors"
	"fmt"
	lock "github.com/bsm/redis-lock"
	"log"
	"net"
	"pos-proxy/config"
	"pos-proxy/db"
	"time"
)

var conn net.Conn

func SendRequest(data []byte) (string, error) {
	//l, err := LockOpera()
	//if err != nil {
	//	log.Println("Couldn't aquire opera lock", err)
	//	return "", err
	//}
	//defer l.Unlock()
	log.Println("About to send message", string(data))
	stx := []byte{0x02}
	etx := []byte{0x03}
	payload := []byte{}
	payload = append(payload, stx...)
	payload = append(payload, data...)
	payload = append(payload, etx...)
	conn.Write(payload)
	message, err := bufio.NewReader(conn).ReadString(etx[0])
	return message, err
}

func Connect() {
	var err error
	connectionString := fmt.Sprintf("%s:5010", config.Config.OperaIP)
	retries := 0
	connected := false
	log.Println("retries", retries, "connected", connected)
	for retries < 3 && !connected {
		log.Println("retries", retries, "connected", connected)
		conn, err = net.Dial("tcp", connectionString)
		if err != nil {
			log.Println("Couldn't connect to opera with err ", err)
			log.Println("Connection string", connectionString)
		} else {
			connected = true
		}
		retries += 1
		time.Sleep(1000 * time.Millisecond)
	}
	if !connected {
		log.Fatal("Couldn't connect to opera on ", connectionString)
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

func LockOpera() (*lock.Lock, error) {
	lockOptions := &lock.LockOptions{
		WaitTimeout: 4 * time.Second,
	}

	l, err := lock.ObtainLock(db.Redis, "opera", lockOptions)
	if err != nil {
		return &lock.Lock{}, err
	} else if l == nil {
		return &lock.Lock{}, errors.New("couldn't obtain opera lock")
	}

	ok, err := l.Lock()
	if err != nil {
		return &lock.Lock{}, err
	} else if !ok {
		return &lock.Lock{}, errors.New("failed to acquire opera lock")
	}

	return l, nil
}
