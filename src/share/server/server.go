package server

import (
	"log"
	"sync"
)

func Run(wg *sync.WaitGroup, port int) {
	defer wg.Done()
	log.Println("Started server on port", port)
}
