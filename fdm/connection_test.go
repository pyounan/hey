package fdm

import (
	"log"
	"testing"

	_ "pos-proxy/config"
)

func TestConnection(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// connection to FDM
	_, err := New()
	if err != nil {
		t.Fail()
		log.Fatal(err)
	}
}
