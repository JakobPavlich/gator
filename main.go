package main

import (
	"fmt"
	"os"

	"github.com/JakobPavlich/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

func NewCommands() commands {
	return commands{
		handler: make(map[string]func(*state, command) error),
	}
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Unable to read config file: %v", err)
	}
	var appState state
	appState.cfg = &cfg
	appCommands := NewCommands()
	appCommands.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Not enough arguments provided")
		os.Exit(1)
	}
	commandName := args[1]
	commandArguments := args[2:]
	loginCommand := command{commandName, commandArguments}
	err = appCommands.run(&appState, loginCommand)
	if err != nil {
		fmt.Printf("Error running login command: %v", err)
	}
	// fmt.Printf("Current user: %s\nCurrent DBURL: %s", appState.cfg.CurrentUserName, appState.cfg.DBURL)
}
