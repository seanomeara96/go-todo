package cli

import (
	"flag"
	"fmt"
	"go-todo/internal/services"
)

type cli struct {
	s *services.Service
}

func New(s *services.Service) *cli {
	return &cli{s}
}

func (cli *cli) Execute() error {

	resource := flag.String("resource", "", "todo, user")
	action := flag.String("action", "", "tidy,...")

	flag.Parse()

	switch *resource {
	case "users":
		return cli.UserActions(*action)
	case "todos":
		return cli.TodoActions(*action)
	default:
		return fmt.Errorf("need to supply a valid resource")
	}

}

func (cli *cli) UserActions(action string) error {
	switch action {

	default:
		return fmt.Errorf("Please supply a valid user action")
	}
}

func (cli *cli) TodoActions(action string) error {
	switch action {
	case "clean":
		return cli.s.DeleteUnattributedTodos()
	default:
		return fmt.Errorf("Please supply a valid todo action")
	}
}
