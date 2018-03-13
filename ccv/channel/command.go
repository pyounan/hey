package channel

import "net"

var commandChannel *net.Conn

func GetCommandChannel() (*net.Conn, error) {
	var err error
	if commandChannel == nil {
		commandChannel, err = newCommandChannel(":4102")
	}
	return commandChannel, err
}

func newCommandChannel(port string) (*net.Conn, error) {
	conn, err := net.Dial("tcp", port)
	return &conn, err
}

/*

type CommandChannel struct {
	conn net.Conn
}


func (channel *CommandChannel) Send(req entity.DeviceRequest) error {
	commandMutex.Lock()
	log.Println("sending")
	defer commandMutex.Unlock()
	buf := bytes.NewBufferString("")
	err := xml.NewEncoder(buf).Encode(req)
	if err != nil {
		log.Println(err)
		return err
	}
	b := buf.Bytes()
	_, err = channel.conn.Write(b)
	if err != nil {
		log.Println(err)
		return err
	}
	channel.conn.Write([]byte("\n"))
	return nil
}

func (channel *CommandChannel) Close() {
	channel.conn.Close()
}

func (channel *CommandChannel) FIN() {
	commandMutex.Lock()
	defer commandMutex.Unlock()
	log.Println("Should send FIN")
	ln, err := channel.conn.Write([]byte("FIN\n"))
	if err != nil {
		log.Println(err)
	}
	log.Println("done sending fin", ln)
}
*/
