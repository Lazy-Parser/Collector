package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Listen listen for some Signal, that mean interruption (CTRL + C). It should block main goroutine
// When interruption fire, then ctxCancel is called to stop whole program
func ListenInterruptionAndStop(ctxCancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sigsString := <-sigs
	fmt.Printf("\nReceived signal: %s, shutting down gracefully...\n", sigsString)
	ctxCancel() // stops all processes
	fmt.Println("Stopped ✅")
}
