package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/kerelape/gophkeeper/internal/stack"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/pior/runnable"
)

// CLI is gophkeeper command-line interface.
type CLI struct {
	Gophkeeper  gophkeeper.Gophkeeper
	CommandLine []string
}

var _ runnable.Runnable = (*CLI)(nil)

// Run implements runnable.Runnable.
//
// @todo #3 Implement `store-file <description> <path>` command.
// @todo #3 Implement `store-card <description> <cardholder> <number> <date> <cvv/cvc> command.
// @todo #3 Implement `store-text <description> <content>` command.
// @todo #3 Implement `restore-file <rid>` command.
// @todo #3 Implement `restore-card <rid>` command.
// @todo #3 Implement `restore-text <rid>` command.
func (c *CLI) Run(ctx context.Context) error {
	var commands = map[string]command{
		"register": &registerCommand{
			gophkeeper: c.Gophkeeper,
		},
		"list": &listCommand{
			gophkeeper: c.Gophkeeper,
		},
		"store-credential": &storeCredentialCommand{
			gophkeeper: c.Gophkeeper,
		},
		"restore-credential": &restoreCredentialCommand{
			gophkeeper: c.Gophkeeper,
		},
	}
	if len(c.CommandLine) < 1 {
		return errors.New("command not specified")
	}
	if c.CommandLine[0] == "help" {
		for n, c := range commands {
			fmt.Printf("%s %s - %s\n", n, c.Help(), c.Description())
		}
		return nil
	}
	if command, ok := commands[c.CommandLine[0]]; ok {
		var (
			input = (stack.Stack[string])(c.CommandLine[1:])
			args  = make(stack.Stack[string], 0, len(input))
		)
		for len(input) > 0 {
			args.Push(input.Pop())
		}
		correct, err := command.Execute(ctx, args)
		if !correct {
			fmt.Printf("%s %s\n", c.CommandLine[0], command.Help())
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to execute command %s: %w", c.CommandLine[0], err)
		}
		return nil
	}
	return errors.New("command not found")
}
