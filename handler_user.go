package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/JakobPavlich/gator/internal/config"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		fmt.Println("username required")
		os.Exit(1)
	}
	cfg, err := config.Read()
	if err != nil {
		return errors.New("error retrieving config file")
	}
	cfg.SetUser(cmd.arguments[0])
	s.cfg.CurrentUserName = cfg.CurrentUserName
	s.cfg.DBURL = cfg.DBURL
	fmt.Printf("User '%s' has been set \n", cfg.CurrentUserName)

	return nil
}
