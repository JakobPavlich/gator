package main

import (
	"fmt"

	"github.com/JakobPavlich/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error retrieving config file: %v", err)
	}
	cfg.SetUser("jakob")
	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("Error retrieving config file: %v", err)
	}
	fmt.Printf("Config file: %+v", cfg)
}
