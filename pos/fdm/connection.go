package fdm

import (
	"errors"
	"strconv"
	"time"

	"github.com/tarm/serial"

	"pos-proxy/config"
	"pos-proxy/libs/libfdm"
)

func Connect(RCRS string) (*libfdm.FDM, error) {
	if RCRS == "" {
		return nil, errors.New("You must specifiy a valid RCRS number")
	}

	// find the FDM that is supposed to receive requests from this RCRS number
	var c *serial.Config
	for _, f := range config.Config.FDMs {
		if f.RCRS == RCRS {
			baudSpeed, _ := strconv.Atoi(f.BaudSpeed)
			c = &serial.Config{Name: f.FDM_Port, Baud: baudSpeed, ReadTimeout: time.Second * 6}
			break
		}
	}

	if c == nil {
		return nil, errors.New("there is no fdm configuration for this production number")
	}

	fdm, err := libfdm.New(c)
	if err != nil {
		return nil, err
	}
	return fdm, nil
}
