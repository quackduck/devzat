package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	chromastyles "github.com/alecthomas/chroma/styles"
	"github.com/jwalton/gchalk"
	markdown "github.com/quackduck/go-term-markdown"
)

func makeFlag(colors []string) func(a string) string {
	flag := make([]*gchalk.Builder, len(colors))
	for i := range colors {
		flag[i] = Chalk.WithHex(colors[i])
	}
	return func(a string) string {
		return applyRainbow(flag, a)
	}
}

func applyRainbow(rainbow []*gchalk.Builder, a string) string {
	a = stripansi.Strip(a)
	buf := ""
	colorOffset := rand.Intn(len(rainbow))
	for i, r := range []rune(a) {
		buf += rainbow[(colorOffset+i)%len(rainbow)].Paint(string(r))
	}
	return buf
}

var (
	Chalk   = gchalk.New(gchalk.ForceLevel(gchalk.LevelAnsi256))
	Green   = ansi256(1, 5, 1)
	Red     = ansi256(5, 1, 1)
	Cyan    = ansi256(1, 5, 5)
	Magenta = ansi256(5, 1, 5)
	Yellow  = ansi256(5, 5, 1)
	Orange  = ansi256(5, 3, 0)
	Blue    = ansi256(0, 3, 5)
	White   = ansi256(5, 5, 5)
	Styles  = []*Style{
		{"white", buildStyle(White)},
		{"red", buildStyle(Red)},
		{"coral", buildStyle(ansi256(5, 2, 2))},
		{"green", buildStyle(Green)},
		{"sky", buildStyle(ansi256(3, 5, 5))},
		{"cyan", buildStyle(Cyan)},
		{"magenta", buildStyle(Magenta)},
		{"pink", buildStyle(ansi256(5, 3, 4))},
		{"rose", buildStyle(ansi256(5, 0, 2))},
		{"cranberry", buildStyle(ansi256(3, 0, 1))},
		{"lavender", buildStyle(ansi256(4, 2, 5))},
		{"fire", buildStyle(ansi256(5, 2, 0))},
		{"pastel green", buildStyle(ansi256(0, 5, 3))},
		{"olive", buildStyle(ansi256(4, 5, 1))},
		{"yellow", buildStyle(Yellow)},
		{"orange", buildStyle(Orange)},
		{"blue", buildStyle(Blue)}}
	SecretStyles = []*Style{
		{"ukraine", buildStyle(Chalk.WithHex("#005bbb").WithBgHex("#ffd500"))},
		{"easter", buildStyle(Chalk.WithRGB(255, 51, 255).WithBgRGB(255, 255, 0))},
		{"baby", buildStyle(Chalk.WithRGB(255, 51, 255).WithBgRGB(102, 102, 255))},
		{"hacker", buildStyle(Chalk.WithRGB(0, 255, 0).WithBgRGB(0, 0, 0))},
		{"l33t", buildStyleNoStrip(Chalk.WithBgBrightBlack())},
		{"whiten", buildStyleNoStrip(Chalk.WithBgWhite())},
		{"trans", makeFlag([]string{"#55CDFC", "#F7A8B8", "#FFFFFF", "#F7A8B8", "#55CDFC"})},
		{"gay", makeFlag([]string{"#FF0018", "#FFA52C", "#FFFF41", "#008018", "#0000F9", "#86007D"})},
		{"lesbian", makeFlag([]string{"#D62E02", "#FD9855", "#FFFFFF", "#D161A2", "#A20160"})},
		{"bi", makeFlag([]string{"#D60270", "#D60270", "#9B4F96", "#0038A8", "#0038A8"})},
		{"ace", makeFlag([]string{"#333333", "#A4A4A4", "#FFFFFF", "#810081"})},
		{"pan", makeFlag([]string{"#FF1B8D", "#FFDA00", "#1BB3FF"})},
		{"enby", makeFlag([]string{"#FFF430", "#FFFFFF", "#9C59D1", "#000000"})},
		{"aro", makeFlag([]string{"#3AA63F", "#A8D47A", "#FFFFFF", "#AAAAAA", "#000000"})},
		{"genderfluid", makeFlag([]string{"#FE75A1", "#FFFFFF", "#BE18D6", "#333333", "#333EBC"})},
		{"agender", makeFlag([]string{"#333333", "#BCC5C6", "#FFFFFF", "#B5F582", "#FFFFFF", "#BCC5C6", "#333333"})},
		{"rainbow", func(a string) string {
			rainbow := []*gchalk.Builder{Red, Orange, Yellow, Green, Cyan, Blue, ansi256(2, 2, 5), Magenta}
			return applyRainbow(rainbow, a)
		}}}
)

func init() {
	markdown.CurrentTheme = chromastyles.ParaisoDark
}

type Style struct {
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
	return Chalk.WithRGB(255/5*r, 255/5*g, 255/5*b)
}

func bgAnsi256(r, g, b uint8) *gchalk.Builder {
	return Chalk.WithBgRGB(255/5*r, 255/5*g, 255/5*b)
}

// Applies color from name
func (u *User) changeColor(colorName string) error {
	style, err := getStyle(colorName)
	if err != nil {
		return err
	}
	if strings.HasPrefix(colorName, "bg-") {
		u.ColorBG = style.name // update bg color
	} else {
		u.Color = style.name // update fg color
	}

	//if colorName == "random" {
	//	u.room.broadcast("", "You're now using "+u.color)
	//}

	u.Name, _ = applyColorToData(u.Name, u.Color, u.ColorBG) // error can be discarded as it has already been checked earlier

	u.term.SetPrompt(u.Name + ": ")
	return nil
}

func applyColorToData(data string, color string, colorBG string) (string, error) {
	styleFG, err := getStyle(color)
	if err != nil {
		return "", err
	}
	styleBG, err := getStyle(colorBG)
	if err != nil {
		return "", err
	}
	return styleBG.apply(styleFG.apply(data)), nil // fg clears the bg color
}

// Sets either the foreground or the background with a random color if the
// given name is correct.
func getRandomColor(name string) *Style {
	var foreground bool
	if name == "random" {
		foreground = true
	} else if name == "bg-random" {
		foreground = false
	} else {
		return nil
	}
	r := rand.Intn(6)
	g := rand.Intn(6)
	b := rand.Intn(6)
	if foreground {
		return &Style{fmt.Sprintf("%03d", r*100+g*10+b), buildStyle(ansi256(uint8(r), uint8(g), uint8(b)))}
	}
	return &Style{fmt.Sprintf("bg-%03d", r*100+g*10+b), buildStyleNoStrip(bgAnsi256(uint8(r), uint8(g), uint8(b)))}
}

// If the input is a named style, returns it. Otherwise, returns nil.
func getNamedColor(name string) *Style {
	for i := range Styles {
		if Styles[i].name == name {
			return Styles[i]
		}
	}
	for i := range SecretStyles {
		if SecretStyles[i].name == name {
			return SecretStyles[i]
		}
	}
	return nil
}

func getCustomColor(name string) (*Style, error) {
	if strings.HasPrefix(name, "#") {
		return &Style{name, buildStyle(Chalk.WithHex(name))}, nil
	}
	if strings.HasPrefix(name, "bg-#") {
		return &Style{name, buildStyleNoStrip(Chalk.WithBgHex(strings.TrimPrefix(name, "bg-")))}, nil
	}
	if len(name) == 3 || len(name) == 6 {
		rgbCode := name
		if strings.HasPrefix(name, "bg-") {
			rgbCode = strings.TrimPrefix(rgbCode, "bg-")
		}
		a, err := strconv.Atoi(rgbCode)
		if err == nil {
			r := (a / 100) % 10
			g := (a / 10) % 10
			b := a % 10
			if r > 5 || g > 5 || b > 5 || r < 0 || g < 0 || b < 0 {
				return nil, errors.New("custom colors have values from 0 to 5 smh")
			}
			if strings.HasPrefix(name, "bg-") {
				return &Style{name, buildStyleNoStrip(bgAnsi256(uint8(r), uint8(g), uint8(b)))}, nil
			}
			return &Style{name, buildStyle(ansi256(uint8(r), uint8(g), uint8(b)))}, nil
		}
		return nil, err
	}
	return nil, nil
}

// Turns name into a style (defaults to nil)
func getStyle(name string) (*Style, error) {
	randomColor := getRandomColor(name)
	if randomColor != nil {
		return randomColor, nil
	}
	if name == "bg-off" {
		return &Style{"bg-off", func(a string) string { return a }}, nil // Used to remove one's background
	}
	namedColor := getNamedColor(name)
	if namedColor != nil {
		return namedColor, nil
	}
	if strings.HasPrefix(name, "#") {
		return &Style{name, buildStyle(Chalk.WithHex(name))}, nil
	}
	customColor, err := getCustomColor(name)
	if err != nil {
		return nil, err
	}
	if customColor != nil {
		return customColor, nil
	}
	//s, err := Chalk.WithStyle(strings.Split(name, "-")...)
	//if err == nil {
	//	return &style{name, buildStyle(s)}, nil
	//}

	return nil, errors.New("Which color? Choose from random, " + strings.Join(func() []string {
		colors := make([]string, 0, len(Styles))
		for i := range Styles {
			colors = append(colors, Styles[i].name)
		}
		return colors
	}(), ", ") + "  \nMake your own colors using hex (#A0FFFF, etc) or RGB values from 0 to 5 (for example, `color 530`, a pretty nice orange). Set bg color like this: color bg-530; remove bg color with color bg-off.\nThere's also a few secret colors :)")
}
