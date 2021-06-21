package main

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/jwalton/gchalk"
)

var (
	chalk   = gchalk.New(gchalk.ForceLevel(gchalk.LevelAnsi256))
	green   = ansi256(1, 5, 1)
	red     = ansi256(5, 1, 1)
	cyan    = ansi256(1, 5, 5)
	magenta = ansi256(5, 1, 3)
	yellow  = ansi256(5, 5, 1)
	orange  = ansi256(5, 3, 0)
	blue    = ansi256(0, 3, 5)
	black   = ansi256(0, 0, 0)
	white   = ansi256(5, 5, 5)
	styles  = []*style{
		{"white", buildStyle(white)},
		{"red", buildStyle(red)},
		{"coral", buildStyle(ansi256(5, 2, 2))},
		{"green", buildStyle(green)},
		{"sky", buildStyle(ansi256(3, 5, 5))},
		{"cyan", buildStyle(cyan)},
		{"magenta", buildStyle(magenta)},
		{"pink", buildStyle(ansi256(5, 3, 4))},
		{"rose", buildStyle(ansi256(5, 0, 2))},
		{"lavender", buildStyle(ansi256(4, 2, 5))},
		{"fire", buildStyle(ansi256(5, 2, 0))},
		{"pastel green", buildStyle(ansi256(0, 5, 3))},
		{"olive", buildStyle(ansi256(4, 5, 1))},
		{"yellow", buildStyle(yellow)},
		{"orange", buildStyle(orange)},
		{"blue", buildStyle(blue)}}
	secretStyles = []*style{
		{"easter", buildStyle(chalk.WithRGB(255, 51, 255).WithBgRGB(255, 255, 0))},
		{"baby", buildStyle(chalk.WithRGB(255, 51, 255).WithBgRGB(102, 102, 255))},
		{"hacker", buildStyle(chalk.WithRGB(0, 255, 0).WithBgRGB(0, 0, 0))},
		{"l33t", buildStyleNoStrip(chalk.WithBgBrightBlack())},
		{"whiten", buildStyleNoStrip(chalk.WithBgWhite())},
		{"rainbow", func(a string) string {
			rainbow := []*gchalk.Builder{red, orange, yellow, green, cyan, blue, magenta}
			a = stripansi.Strip(a)
			buf := ""
			colorOffset := rand.Intn(len(rainbow))
			for i := range []rune(a) {
				buf += rainbow[(colorOffset+i)%len(rainbow)].Paint(string(a[i]))
			}
			return buf
		}}}
)

type style struct {
	name  string
	apply func(string) string
}

func buildStyle(c *gchalk.Builder) func(string) string {
	return func(s string) string {
		return c.Paint(stripansi.Strip(s))
	}
}

func buildStyleNoStrip(c *gchalk.Builder) func(string) string {
	return func(s string) string {
		return c.Paint(s)
	}
}

// with r, g and b values from 0 to 5
func ansi256(r, g, b uint8) *gchalk.Builder {
	return chalk.WithRGB(255/5*r, 255/5*g, 255/5*b)
}

func bgAnsi256(r, g, b uint8) *gchalk.Builder {
	return chalk.WithBgRGB(255/5*r, 255/5*g, 255/5*b)
}

// Applies color from name
func (u *user) changeColor(colorName string) error {
	style, err := getStyle(colorName)
	if err != nil {
		return err
	}
	u.color = style.name
	u.name = style.apply(u.name)
	u.term.SetPrompt(u.name + ": ")
	if _, ok := allUsers[u.id]; !ok {
		mainRoom.broadcast(devbot, "You seem to be new here "+u.name+". Welcome to Devzat! Run ./help to see what you can do.")
	}
	allUsersMutex.Lock()
	allUsers[u.id] = u.name
	allUsersMutex.Unlock()
	saveBansAndUsers()
	return nil
}

// Turns name into a style (defaults to nil)
func getStyle(name string) (*style, error) {
	if name == "random" {
		r := rand.Intn(6)
		g := rand.Intn(6)
		b := rand.Intn(6)
		return &style{strconv.Itoa(r*100 + g*10 + b), buildStyle(ansi256(uint8(r), uint8(g), uint8(b)))}, nil
	}
	for i := range styles {
		if styles[i].name == name {
			return styles[i], nil
		}
	}
	for i := range secretStyles {
		if secretStyles[i].name == name {
			return secretStyles[i], nil
		}
	}
	if strings.HasPrefix(name, "#") {
		return &style{name, buildStyle(chalk.WithHex(name))}, nil
	}
	if len(name) == 3 || len(name) == 5 {
		rgbCode := name
		if strings.HasPrefix(name, "bg") {
			rgbCode = strings.TrimPrefix(rgbCode, "bg")
		}
		a, err := strconv.Atoi(rgbCode)
		if err == nil {
			r := (a / 100) % 10
			g := (a / 10) % 10
			b := a % 10
			if r > 5 || g > 5 || b > 5 || r < 0 || g < 0 || b < 0 {
				return nil, errors.New("custom colors have values from 0 to 5 smh")
			}
			if strings.HasPrefix(name, "bg") {
				return &style{name, buildStyleNoStrip(bgAnsi256(uint8(r), uint8(g), uint8(b)))}, nil
			}
			return &style{name, buildStyle(ansi256(uint8(r), uint8(g), uint8(b)))}, nil
		}
	}
	return nil, errors.New("Which color? Choose from random, " + strings.Join(func() []string {
		colors := make([]string, 0, len(styles))
		for i := range styles {
			colors = append(colors, styles[i].name)
		}
		return colors
	}(), ", ") + "  \nMake your own colors using hex (#A0FFFF, etc) or RGB values from 0 to 5 (for example, ./color 530, a pretty nice orange). Set bg color like this: ./color bg530.  \nThere's also a few secret colors :)")
}
