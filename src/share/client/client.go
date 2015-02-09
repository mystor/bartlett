package client

import (
	"log"
	"sync"
)

func Run(wg *sync.WaitGroup, url string, localPort int) {
	defer wg.Done()

	log.Println("Started client, connecting to", url, "(localPort:", localPort, ")")
}
