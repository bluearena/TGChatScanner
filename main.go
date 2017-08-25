package main

import (
	"github.com/zwirec/TGChatScanner/service"
	"log"
)

func main() {
	serv := service.NewService()
	if err := serv.Run(); err != nil {
		log.Fatal(err)
	}

}
