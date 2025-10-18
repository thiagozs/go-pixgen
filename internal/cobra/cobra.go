package cobra

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Command is a lightweight substitute for github.com/spf13/cobra.Command
// implementing only the features used by this project.
type Command struct {
	Use   string
	Short string
	Long  string

	Run  func(cmd *Command, args []string)
	RunE func(cmd *Command, args []string) error

	SilenceUsage  bool
	SilenceErrors bool

	flagSet             *flag.FlagSet
	subCommands         []*Command
	parent              *Command
	helpRequested       bool
	helpRequestedShort  bool
	usageTemplateWriter io.Writer
}

// Execute finds and executes the appropriate command using os.Args.
func (c *Command) Execute() error {
	return c.execute(os.Args[1:])
}

func (c *Command) execute(args []string) error {
	if len(args) > 0 {
		if sub := c.findSubCommand(args[0]); sub != nil {
			return sub.execute(args[1:])
		}
	}

	fs := c.Flags()
	if fs != nil {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				c.Usage()
				return nil
			}
			if c.SilenceUsage {
				return err
			}
			c.Usage()
			return err
		}

		if c.helpRequested || c.helpRequestedShort {
			c.Usage()
			return nil
		}

		args = fs.Args()
	}

	if c.RunE != nil {
		if err := c.RunE(c, args); err != nil {
			if c.SilenceErrors {
				return err
			}
			return err
		}
		return nil
	}

	if c.Run != nil {
		c.Run(c, args)
		return nil
	}

	if len(c.subCommands) > 0 {
		if !c.SilenceUsage {
			c.Usage()
		}
		if c.SilenceErrors {
			return nil
		}
		return errors.New("command requires a subcommand")
	}

	return nil
}

// Flags returns the flag set for the command, instantiating it if necessary.
func (c *Command) Flags() *flag.FlagSet {
	if c.flagSet == nil {
		name := c.Name()
		if name == "" {
			name = "root"
		}
		c.flagSet = flag.NewFlagSet(name, flag.ContinueOnError)
		c.flagSet.SetOutput(io.Discard)
		c.flagSet.BoolVar(&c.helpRequested, "help", false, "display help")
		c.flagSet.BoolVar(&c.helpRequestedShort, "h", false, "display help")
	}
	return c.flagSet
}

// AddCommand registers subcommands.
func (c *Command) AddCommand(cmds ...*Command) {
	for _, cmd := range cmds {
		cmd.parent = c
		c.subCommands = append(c.subCommands, cmd)
	}
}

// Usage prints the command usage information.
func (c *Command) Usage() {
	writer := c.usageWriter()
	if c.Long != "" {
		fmt.Fprintln(writer, c.Long)
	} else if c.Short != "" {
		fmt.Fprintln(writer, c.Short)
	}

	fmt.Fprintf(writer, "Usage: %s\n", c.fullUseLine())

	if len(c.subCommands) > 0 {
		fmt.Fprintln(writer, "\nCommands:")
		for _, sub := range c.subCommands {
			fmt.Fprintf(writer, "  %-12s %s\n", sub.Name(), sub.Short)
		}
	}

	if c.flagSet != nil && c.flagSet.NFlag() > 0 {
		fmt.Fprintln(writer, "\nFlags:")
		c.flagSet.SetOutput(writer)
		c.flagSet.PrintDefaults()
	}
}

func (c *Command) usageWriter() io.Writer {
	if c.usageTemplateWriter != nil {
		return c.usageTemplateWriter
	}
	return os.Stderr
}

// Name returns the first word of the Use field.
func (c *Command) Name() string {
	if c.Use == "" {
		return ""
	}
	fields := strings.Fields(c.Use)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func (c *Command) findSubCommand(name string) *Command {
	for _, sub := range c.subCommands {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}

func (c *Command) fullUseLine() string {
	segments := []string{}
	for cmd := c; cmd != nil; cmd = cmd.parent {
		segments = append([]string{cmd.Name()}, segments...)
	}
	return strings.TrimSpace(strings.Join(segments, " "))
}
