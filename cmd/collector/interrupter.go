package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Listen listen to some Signal, that mean interruption (CTRL + C). Should not call in goroutine, because it should block main goroutine
// When interruption fire, then ctxCancel calls to stop all program
func ListenInterruptionAndStop(ctxCancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sigsString := <-sigs
	fmt.Printf("\nReceived signal: %s, shutting down gracefully...\n", sigsString)
	ctxCancel() // stops all processes
	fmt.Println("Stopped ✅")
}
