package main

import (
	"log"

	"github.com/lavab/kiri"
)

func main() {
	sd := kiri.New([]string{
		"http://10.10.20.49:4001",
		"http://10.10.20.50:4001",
		"http://10.10.20.51:4001",
	})
	sd.Store(kiri.Default, "/kiri")

	matched, err := sd.Query("api-master", nil)
	if err != nil {
		log.Print(err)
	}

	for _, match := range matched {
		log.Print(match.Address)
	}

	select {}
}
