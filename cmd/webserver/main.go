package main

import (
	"fmt"
	"log"
	"os"

	"github.com/civ13/ycom/web"
)

func main() {
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	log.Printf("YCOM Web Server starting on %s", addr)
	log.Printf("Open http://localhost%s in your browser", addr)

	// Start web server
	web.StartServer(addr)

	// Keep server running
	fmt.Println("Press Ctrl+C to stop")
	select {}
}
