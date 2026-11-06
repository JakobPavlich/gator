package main

import "errors"

type command struct {
	name      string
	arguments []string
}

type commands struct {
	handler map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, contains := c.handler[cmd.name]
	if contains {
		err := f(s, cmd)
		if err != nil {
			return errors.New(err.Error())
		}
		return nil
	}
	return errors.New("command not found")
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handler[name] = f
}
