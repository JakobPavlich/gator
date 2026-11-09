package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/JakobPavlich/gator/internal/config"
	"github.com/JakobPavlich/gator/internal/database"
	"github.com/JakobPavlich/gator/internal/rss"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		fmt.Println("username required")
		os.Exit(1)
	}
	userName := cmd.arguments[0]
	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		fmt.Printf("User '%s' is not yet registered. Use the 'register (username)' command first.", userName)
		os.Exit(1)
	}
	cfg, err := config.Read()
	if err != nil {
		return errors.New("error retrieving config file")
	}
	cfg.SetUser(userName)
	s.cfg.CurrentUserName = cfg.CurrentUserName
	s.cfg.DBURL = cfg.DBURL
	fmt.Printf("User '%s' has been set \n", cfg.CurrentUserName)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		fmt.Println("username required")
		os.Exit(1)
	}
	userName := cmd.arguments[0]

	userParams := database.CreateUserParams{}
	userParams.ID = uuid.New()
	userParams.CreatedAt = time.Now()
	userParams.UpdatedAt = time.Now()
	user, err := s.db.GetUser(context.Background(), userName)
	if err == nil {
		fmt.Printf("User '%s' already exists", user.Name)
		os.Exit(1)
	}
	userParams.Name = userName
	user, err = s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		return fmt.Errorf("couldn't create a user: %v", err)
	}
	s.cfg.SetUser(user.Name)
	fmt.Println("User was successfully created.")
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("Data: %+v", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DropTable(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Database was reset successfully")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	_, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	return nil
}
