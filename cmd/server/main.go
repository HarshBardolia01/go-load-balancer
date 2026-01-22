package main

import (
	"context"
	"fmt"
	"go-load-balancer/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFile, err := server.ParseArguments()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	config, err := server.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := server.GetServer(config)

	if err := s.Run(ctx); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
