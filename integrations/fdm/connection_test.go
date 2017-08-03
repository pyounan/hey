package fdm

import (
	"testing"

	_ "pos-proxy/config"
	"pos-proxy/fdm"
)

func TestConnection(t *testing.T) {
	// connection to FDM
	_, err := fdm.New()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
}

func TestCheckStatus(t *testing.T) {
	FDM, err := fdm.New()
	if err != nil {
		t.Fail()
		t.Log(err)
	}

	res, err := FDM.CheckStatus()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	if res != true {
		t.Fail()
		t.Log("FDM is offline")
	}
}
