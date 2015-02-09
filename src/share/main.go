package main

import (
	"flag"
	"fmt"
	"log"
)

type options struct {
	noclient  bool   // Should we run a client (inverted)
	server    bool   // Do we want to run a server at all
	port      int    // Is unused by the client
	localPort int    // Port used locally for editors
	url       string // "localhost:port" means the local server which we are hosting
}

func parseFlags() options {
	var opts options

	flag.BoolVar(&opts.noclient, "noclient", false,
		"Don't run the client and synchronise the current directory.")
	flag.IntVar(&opts.port, "port", 3000,
		"The port to run the server on (no effect when connecting to server)")
	flag.IntVar(&opts.localPort, "localPort", 8232,
		"The port to use to locally communicate with editors. Probably don't change...")
	flag.Parse()

	if flag.NArg() > 1 {
		log.Fatal("Can only connect to one server")
	}

	if flag.NArg() == 1 {
		opts.server = false
		opts.url = flag.Arg(0)
	} else if flag.NArg() == 0 {
		opts.server = true
		opts.url = "localhost:" + fmt.Sprintf("%d", opts.port)
	}

	return opts
}

func main() {

	fmt.Println(parseFlags())
}
