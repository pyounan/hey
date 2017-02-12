package main

import (
	"pos-proxy/fdm"
	_ "pos-proxy/config"
)


func main() {
	// connection to FDM
	fdm.New()
}
