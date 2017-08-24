package main

import (
	"github.com/zwirec/TGChatScanner/service"
	"log"
	"github.com/zwirec/TGChatScanner/modelManager"
)

func main() {
	serv := service.NewService()
	if err := serv.Run(); err != nil {
		log.Fatal(err)
	}

}
