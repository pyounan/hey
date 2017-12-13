package logging

import (
	"fmt"
	"log"
	"time"
)

func Info(str string) {
	formatted := fmt.Sprintf("INFO: %s", str)
	log.Println("=========================")
	log.Println("TimeStamp", time.Now())
	log.Println(formatted)
	log.Println("=========================")
}

func Alert(str string) {
	formatted := fmt.Sprintf("ALERT: %s", str)
	log.Println("=========================")
	log.Println("TimeStamp", time.Now())
	log.Println(formatted)
	log.Println("=========================")
}

func Error(str string) {
	formatted := fmt.Sprintf("ERROR: %s", str)
	log.Println("=========================")
	log.Println("TimeStamp", time.Now())
	log.Println(formatted)
	log.Println("=========================")
}

func Debug(str string) {
	formatted := fmt.Sprintf("DEBUG: %s", str)
	log.Println("=========================")
	log.Println("TimeStamp", time.Now())
	log.Println(formatted)
	log.Println("=========================")
}
