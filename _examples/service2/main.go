package main

import (
	"github.com/lavab/kiri"
)

func main() {
	sd := kiri.New([]string{
		"http://10.10.20.49:4001",
		"http://10.10.20.50:4001",
		"http://10.10.20.51:4001",
	})
	sd.Register("api-master", "10.10.100.10:8001", nil)
	sd.Store(kiri.Default, "/kiri")
	select {}
}
