package cmd

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = ""
	argsInfo = ""
	info     = ""
)

type Command struct{}

func (c *Command) Name() string {
	return name
}

func (c *Command) ArgsInfo() string {
	return argsInfo
}

func (c *Command) Info() string {
	return info
}

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisNormal
}

func (c *Command) Fn(_ string, u i.User) error {
	byVisibility := make(map[models.CommandVisibility][]i.Command)

	for _, cmd := range u.Room().Server().Commands() {
		v := cmd.Visibility()
		byVisibility[v] = append(byVisibility[v], cmd)
	}

	if len(byVisibility[models.CommandVisNormal]) > 0 {
		strCommands := autogenCommands(byVisibility[models.CommandVisNormal])
		msg := fmt.Sprintf("commands:\n%s\n\n", strCommands)
		u.Room().Broadcast("", msg)
	}

	if len(byVisibility[models.CommandVisLow]) > 0 {
		strCommands := autogenCommands(byVisibility[models.CommandVisLow])
		msg := fmt.Sprintf("commands:\n%s\n\n", strCommands)
		u.Room().Broadcast("", msg)
	}

	if !u.IsAdmin() {
		return nil
	}

	return nil
}

func autogenCommands(cmds []i.Command) string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)

	for _, c := range cmds {
		formatted := fmt.Sprintf("   %s\t%s\t_%s_  \n", c.Name(), c.ArgsInfo(), c.Info())
		_, _ = w.Write([]byte(formatted))
	}

	_ = w.Flush()

	return b.String()
}
