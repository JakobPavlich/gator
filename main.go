package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/JakobPavlich/gator/internal/config"
	"github.com/JakobPavlich/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func NewCommands() commands {
	return commands{
		handler: make(map[string]func(*state, command) error),
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("could not find user '%s'", user.Name)
		}
		err = handler(s, c, user)
		if err != nil {
			return err
		}
		return nil
	}
}

func main() {
	// Nastavljanje uporabnika
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Unable to read config file: %v", err)
	}
	var appState state
	appState.cfg = &cfg
	// SQL
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Printf("Could not connect to the database: %v", err)
		os.Exit(1)
	}
	appState.db = database.New(db)

	appCommands := NewCommands()
	appCommands.register("login", handlerLogin)
	appCommands.register("register", handlerRegister)
	appCommands.register("reset", handlerReset)
	appCommands.register("users", handlerUsers)
	appCommands.register("agg", handlerAgg)
	appCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	appCommands.register("feeds", handlerFeeds)
	appCommands.register("follow", middlewareLoggedIn(handlerFollow))
	appCommands.register("following", middlewareLoggedIn(handlerFollowing))
	appCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Not enough arguments provided")
		os.Exit(1)
	}
	commandName := args[1]
	commandArguments := args[2:]
	command := command{commandName, commandArguments}
	err = appCommands.run(&appState, command)
	if err != nil {
		fmt.Printf("Error running command: %v", err)
		os.Exit(1)
	}
}
