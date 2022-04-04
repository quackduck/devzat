package colors

import (
	"errors"
	"fmt"
	"github.com/alecthomas/chroma"
	"math/rand"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	chromastyles "github.com/alecthomas/chroma/styles"
	"github.com/jwalton/gchalk"
	markdown "github.com/quackduck/go-term-markdown"
)

func NewFormatter() *Formatter {
	f := &Formatter{}

	f.Init()

	return f
}

type Formatter struct {
	chalk  *gchalk.Builder
	colors Colors
	Styles struct {
		Normal []*Style
		Secret []*Style
	}
}

func (f *Formatter) Init() {
	f.chalk = gchalk.New(gchalk.ForceLevel(gchalk.LevelAnsi256))
	f.colors.init()
	markdown.CurrentTheme = chromastyles.ParaisoDark
	f.initChromaStyles()

	f.Styles.Normal = []*Style{
		{White, f.buildStyle(f.colors.White)},
		{Red, f.buildStyle(f.colors.Red)},
		{Coral, f.buildStyle(f.Ansi256(5, 2, 2))},
		{Green, f.buildStyle(f.colors.Green)},
		{Sky, f.buildStyle(f.Ansi256(3, 5, 5))},
		{Cyan, f.buildStyle(f.colors.Cyan)},
		{Magenta, f.buildStyle(f.colors.Magenta)},
		{Pink, f.buildStyle(f.Ansi256(5, 3, 4))},
		{Rose, f.buildStyle(f.Ansi256(5, 0, 2))},
		{Cranberry, f.buildStyle(f.Ansi256(3, 0, 1))},
		{Lavender, f.buildStyle(f.Ansi256(4, 2, 5))},
		{Fire, f.buildStyle(f.Ansi256(5, 2, 0))},
		{PastelGreen, f.buildStyle(f.Ansi256(0, 5, 3))},
		{Olive, f.buildStyle(f.Ansi256(4, 5, 1))},
		{Yellow, f.buildStyle(f.colors.Yellow)},
		{Orange, f.buildStyle(f.colors.Orange)},
		{Blue, f.buildStyle(f.colors.Blue)},
	}

	f.Styles.Secret = []*Style{
		{Ukraine, f.buildStyle(f.chalk.WithHex("#005bbb").WithBgHex("#ffd500"))},
		{Easter, f.buildStyle(f.chalk.WithRGB(255, 51, 255).WithBgRGB(255, 255, 0))},
		{Baby, f.buildStyle(f.chalk.WithRGB(255, 51, 255).WithBgRGB(102, 102, 255))},
		{Hacker, f.buildStyle(f.chalk.WithRGB(0, 255, 0).WithBgRGB(0, 0, 0))},
		{L33t, f.buildStyleNoStrip(f.chalk.WithBgBrightBlack())},
		{Whiten, f.buildStyleNoStrip(f.chalk.WithBgWhite())},
		{Trans, f.makeFlag([]string{"#55CDFC", "#F7A8B8", "#FFFFFF", "#F7A8B8", "#55CDFC"})},
		{Gay, f.makeFlag([]string{"#FF0018", "#FFA52C", "#FFFF41", "#008018", "#0000F9", "#86007D"})},
		{Lesbian, f.makeFlag([]string{"#D62E02", "#FD9855", "#FFFFFF", "#D161A2", "#A20160"})},
		{Bi, f.makeFlag([]string{"#D60270", "#D60270", "#9B4F96", "#0038A8", "#0038A8"})},
		{Ace, f.makeFlag([]string{"#333333", "#A4A4A4", "#FFFFFF", "#810081"})},
		{Pan, f.makeFlag([]string{"#FF1B8D", "#FFDA00", "#1BB3FF"})},
		{Enby, f.makeFlag([]string{"#FFF430", "#FFFFFF", "#9C59D1", "#000000"})},
		{Aro, f.makeFlag([]string{"#3AA63F", "#A8D47A", "#FFFFFF", "#AAAAAA", "#000000"})},
		{Genderfluid, f.makeFlag([]string{"#FE75A1", "#FFFFFF", "#BE18D6", "#333333", "#333EBC"})},
		{Agender, f.makeFlag([]string{"#333333", "#BCC5C6", "#FFFFFF", "#B5F582", "#FFFFFF", "#BCC5C6", "#333333"})},
		{Rainbow, func(a string) string {
			rainbow := []*gchalk.Builder{
				f.colors.Red,
				f.colors.Orange,
				f.colors.Yellow,
				f.colors.Green,
				f.colors.Cyan,
				f.colors.Blue,
				f.Ansi256(2, 2, 5),
				f.colors.Magenta,
			}
			return f.ApplyRainbow(rainbow, a)
		}}}
}

// add Matt Gleich's blackbird theme from:
// https://github.com/blackbirdtheme/vscode/blob/master/themes/blackbird-midnight-color-theme.json#L175
func (f *Formatter) initChromaStyles() {
	red := "#ff1131" // added saturation
	redItalic := "italic " + red
	white := "#fdf7cd"
	yellow := "#e1db3f"
	blue := "#268ef8"  // added saturation
	green := "#22e327" // added saturation
	gray := "#5a637e"
	teal := "#00ecd8"
	tealItalic := "italic " + teal

	chromastyles.Register(chroma.MustNewStyle("blackbird", chroma.StyleEntries{
		chroma.Text:                white,
		chroma.Error:               red,
		chroma.Comment:             gray,
		chroma.Keyword:             redItalic,
		chroma.KeywordNamespace:    redItalic,
		chroma.KeywordType:         tealItalic,
		chroma.Operator:            blue,
		chroma.Punctuation:         white,
		chroma.Name:                white,
		chroma.NameAttribute:       white,
		chroma.NameClass:           green,
		chroma.NameConstant:        tealItalic,
		chroma.NameDecorator:       green,
		chroma.NameException:       red,
		chroma.NameFunction:        green,
		chroma.NameOther:           white,
		chroma.NameTag:             yellow,
		chroma.LiteralNumber:       blue,
		chroma.Literal:             yellow,
		chroma.LiteralDate:         yellow,
		chroma.LiteralString:       yellow,
		chroma.LiteralStringEscape: teal,
		chroma.GenericDeleted:      red,
		chroma.GenericEmph:         "italic",
		chroma.GenericInserted:     green,
		chroma.GenericStrong:       "bold",
		chroma.GenericSubheading:   yellow,
		chroma.Background:          "bg:#000000",
	}))
}

func (f *Formatter) makeFlag(colors []string) func(a string) string {
	flag := make([]*gchalk.Builder, len(colors))
	for i := range colors {
		flag[i] = f.chalk.WithHex(colors[i])
	}
	return func(a string) string {
		return f.ApplyRainbow(flag, a)
	}
}

func (f *Formatter) ApplyRainbow(rainbow []*gchalk.Builder, a string) string {
	a = stripansi.Strip(a)
	buf := ""
	colorOffset := rand.Intn(len(rainbow))
	for i, r := range []rune(a) {
		buf += rainbow[(colorOffset+i)%len(rainbow)].Paint(string(r))
	}
	return buf
}

func (f *Formatter) buildStyle(c *gchalk.Builder) func(string) string {
	return func(s string) string {
		return c.Paint(stripansi.Strip(s))
	}
}

func (f *Formatter) buildStyleNoStrip(c *gchalk.Builder) func(string) string {
	return func(s string) string {
		return c.Paint(s)
	}
}

// with r, g and b values from 0 to 5
func (f *Formatter) Ansi256(r, g, b uint8) *gchalk.Builder {
	return f.chalk.WithRGB(255/5*r, 255/5*g, 255/5*b)
}

func (f *Formatter) BgAnsi256(r, g, b uint8) *gchalk.Builder {
	return f.chalk.WithBgRGB(255/5*r, 255/5*g, 255/5*b)
}

func (f *Formatter) ApplyColorToData(data string, color string, colorBG string) (string, error) {
	styleFG, err := f.GetStyle(color)
	if err != nil {
		return "", err
	}

	styleBG, err := f.GetStyle(colorBG)
	if err != nil {
		return "", err
	}

	return styleBG.Apply(styleFG.Apply(data)), nil // fg clears the bg color
}

// Sets either the foreground or the background with a random color if the
// given name is correct.
func (f *Formatter) GetRandomColor(name string) *Style {
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
		return &Style{fmt.Sprintf("%03d", r*100+g*10+b), f.buildStyle(f.Ansi256(uint8(r), uint8(g), uint8(b)))}
	}

	return &Style{fmt.Sprintf("bg-%03d", r*100+g*10+b), f.buildStyleNoStrip(f.BgAnsi256(uint8(r), uint8(g), uint8(b)))}
}

// If the input is a named style, returns it. Otherwise, returns nil.
func (f *Formatter) GetNamedColor(name string) *Style {
	for _, s := range f.Styles.Normal {
		if s.Name == name {
			return s
		}
	}

	for _, s := range f.Styles.Secret {
		if s.Name == name {
			return s
		}
	}

	return nil
}

func (f *Formatter) GetCustomColor(name string) (*Style, error) {
	if strings.HasPrefix(name, "#") {
		return &Style{name, f.buildStyle(f.chalk.WithHex(name))}, nil
	}

	if strings.HasPrefix(name, "bg-#") {
		return &Style{name, f.buildStyleNoStrip(f.chalk.WithBgHex(strings.TrimPrefix(name, "bg-")))}, nil
	}

	if len(name) != 3 && len(name) != 6 {
		return nil, nil
	}

	rgbCode := name

	if strings.HasPrefix(name, "bg-") {
		rgbCode = strings.TrimPrefix(rgbCode, "bg-")
	}

	a, err := strconv.Atoi(rgbCode)
	if err != nil {
		return nil, err
	}

	r := (a / 100) % 10
	g := (a / 10) % 10
	b := a % 10

	if r > 5 || g > 5 || b > 5 || r < 0 || g < 0 || b < 0 {
		return nil, errors.New("custom colors have values from 0 to 5 smh")
	}

	if strings.HasPrefix(name, "bg-") {
		return &Style{name, f.buildStyleNoStrip(f.BgAnsi256(uint8(r), uint8(g), uint8(b)))}, nil
	}

	return &Style{name, f.buildStyle(f.Ansi256(uint8(r), uint8(g), uint8(b)))}, nil
}

// Turns name into a style (defaults to nil)
func (f *Formatter) GetStyle(name string) (*Style, error) {
	randomColor := f.GetRandomColor(name)
	if randomColor != nil {
		return randomColor, nil
	}
	if name == "bg-off" {
		return &Style{"bg-off", func(a string) string { return a }}, nil // Used to remove one's background
	}

	namedColor := f.GetNamedColor(name)
	if namedColor != nil {
		return namedColor, nil
	}
	if strings.HasPrefix(name, "#") {
		return &Style{name, f.buildStyle(f.chalk.WithHex(name))}, nil
	}

	customColor, err := f.GetCustomColor(name)
	if err != nil {
		return nil, err
	}
	if customColor != nil {
		return customColor, nil
	}
	return nil, errors.New("Which color? Choose from random, " + strings.Join(func() []string {
		colors := make([]string, 0, len(f.Styles.Normal))
		for i := range f.Styles.Normal {
			colors = append(colors, f.Styles.Normal[i].Name)
		}
		return colors
	}(), ", ") + "  \nMake your own colors using hex (#A0FFFF, etc) or RGB values from 0 to 5 (for example, `color 530`, a pretty nice Orange). Set bg color like this: color bg-530; remove bg color with color bg-off.\nThere's also a few secret colors :)")
}

func (f *Formatter) GetStyleNames() []string {
	names := make([]string, 0)
	for _, style := range f.Styles.Normal {
		names = append(names, style.Name)
	}

	return names
}

func (f *Formatter) Colors() Colors {
	return f.colors
}
