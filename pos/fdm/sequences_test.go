package fdm

import (
	"pos-proxy/db"
	"testing"
)

func TestSequenceGeneration(t *testing.T) {
	db.Connect()
	s, err := GetNextSequence("BHES004CLOUD01")
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Logf("coroutine1: sequence %d", s)
	s, err = GetNextSequence("BHES004CLOUD01")
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Logf("coroutine2: sequence %d", s)
}
