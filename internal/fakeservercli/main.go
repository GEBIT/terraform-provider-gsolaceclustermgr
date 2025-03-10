// manual starter for fakeserver (test the provider without accessing the actual solace api).

package main

import (
	"flag"
	"fmt"
	"os"

	fakeserver "terraform-provider-gsolaceclustermgr/internal/fakeserver"
)

func main() {
	apiServerObjects := make(map[string]fakeserver.ServiceInfo)

	port := flag.Int("port", 8091, "The port fakeserver will listen on")
	debug := flag.Bool("debug", false, "Enable debug output of the server")
	baseSid := flag.Int("base-sid", 0, "generate SIDs from seqeunce starting with this. 0 = UUID-Generation instead")

	flag.Parse()

	svr := fakeserver.NewFakeServer(*port, apiServerObjects, false, *debug, *baseSid)

	fmt.Printf("Starting server on port %d...\n", *port)

	internalServer := svr.GetServer()
	err := internalServer.ListenAndServe()
	if nil != err {
		fmt.Printf("Error with the internal TCP server: %s", err)
		os.Exit(1)
	}
}
