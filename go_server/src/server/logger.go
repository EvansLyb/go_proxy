package server

import (
	"log"
	"os"
	"time"
)

var Logger log.Logger



func init(){
	file, err := os.OpenFile("/var/log/transparent_proxy.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		log.Fatal(err)
	}
	Logger.SetOutput(file)
	Logger.SetPrefix(time.Now().String() + "	")
}