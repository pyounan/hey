package channel

import "net"

var deviceChannel *net.Listener

func GetDeviceChannel() (*net.Listener, error) {
	var err error
	if deviceChannel == nil {
		deviceChannel, err = newDeviceChannel(":4100")
	}
	return deviceChannel, err
}

func newDeviceChannel(port string) (*net.Listener, error) {
	conn, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}
