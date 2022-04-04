package colour

import (
	"devzat/pkg/commands/color"
	"devzat/pkg/models"
)

// this is just an alias to the color command with a different name "colour"

type Command struct {
	color.Command
}

func (c *Command) Name() string {
	return "colour"
}

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisHidden
}
