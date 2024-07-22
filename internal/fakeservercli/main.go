// starter for fakeserver, needed to test the provider without accessing the actual solace api (e.g. CI).

package main

import (
	"flag"
	"fmt"
	"os"

	fakeserver "terraform-provider-gsolaceclustermgr/internal/fakeserver"
)

func main() {
	apiServerObjects := make(map[string]fakeserver.ServiceInfo)

	port := flag.Int("port", 8080, "The port fakeserver will listen on")
	debug := flag.Bool("debug", false, "Enable debug output of the server")

	flag.Parse()

	svr := fakeserver.NewFakeServer(*port, apiServerObjects, false, *debug)

	fmt.Printf("Starting server on port %d...\n", *port)

	internalServer := svr.GetServer()
	err := internalServer.ListenAndServe()
	if nil != err {
		fmt.Printf("Error with the internal TCP server: %s", err)
		os.Exit(1)
	}
}
