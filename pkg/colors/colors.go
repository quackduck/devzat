package colors

import "github.com/jwalton/gchalk"

type Colors struct {
	chalk   *gchalk.Builder
	Green   *gchalk.Builder
	Red     *gchalk.Builder
	Cyan    *gchalk.Builder
	Magenta *gchalk.Builder
	Yellow  *gchalk.Builder
	Orange  *gchalk.Builder
	Blue    *gchalk.Builder
	White   *gchalk.Builder
}

func (c *Colors) ansi256(r, g, b uint8) *gchalk.Builder {
	return c.chalk.WithRGB(255/5*r, 255/5*g, 255/5*b)
}

func (c *Colors) init() {
	c.chalk = gchalk.New(gchalk.ForceLevel(gchalk.LevelAnsi256))

	c.Green = c.ansi256(1, 5, 1)
	c.Red = c.ansi256(5, 1, 1)
	c.Cyan = c.ansi256(1, 5, 5)
	c.Magenta = c.ansi256(5, 1, 5)
	c.Yellow = c.ansi256(5, 5, 1)
	c.Orange = c.ansi256(5, 3, 0)
	c.Blue = c.ansi256(0, 3, 5)
	c.White = c.ansi256(5, 5, 5)
}
