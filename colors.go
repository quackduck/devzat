package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	chromastyles "github.com/alecthomas/chroma/styles"
	"github.com/fatih/color"
	markdown "github.com/quackduck/go-term-markdown"
)

var (
    White  = color.New(color.FgWhite)
    Red    = color.New(color.FgRed)
    Green  = color.New(color.FgGreen)
    Cyan   = color.New(color.FgCyan)
    Magenta = color.New(color.FgMagenta)
    Yellow = color.New(color.FgYellow)
    Orange = color.New(color.FgHiYellow)
    Blue   = color.New(color.FgBlue)
    Styles = []*color.Color{
        {"white", White},
        {"red", Red},
        {"coral", color.New(color.FgHiRed)},
        {"green", Green},
        {"sky", color.New(color.FgHiCyan)},
        {"cyan", Cyan},
        {"magenta", Magenta},
        {"pink", color.New(color.FgHiMagenta)},
        {"rose", color.New(color.FgHiRed)},
        {"cranberry", color.New(color.FgHiBlack)},
        {"lavender", color.New(color.FgHiBlack).Add(color.BgMagenta)},
        {"fire", color.New(color.FgHiYellow)},
        {"pastelgreen", color.New(color.FgGreen)},
        {"olive", color.New(color.FgHiBlack).Add(color.BgGreen)},
        {"yellow", Yellow},
        {"orange", Orange},
        {"blue", Blue},
    }
    SecretStyles = []*color.Color{
        {"elitedino", color.New(color.FgHiBlack).Add(color.BgRed)},
        {"ukraine", color.New(color.FgHiBlack).Add(color.BgHiYellow)},
        {"easter", color.New(color.FgMagenta).Add(color.BgYellow)},
        {"baby", color.New(color.FgMagenta).Add(color.BgBlue)},
        {"hacker", color.New(color.FgHiGreen).Add(color.BgBlack)},
        {"l33t", color.New(color.BgHiBlack)},
        {"whiten", color.New(color.BgWhite)},
        {"trans", color.New(color.FgHiBlue).Add(color.BgHiMagenta)},
        {"gay", color.New(color.FgRed).Add(color.BgHiMagenta)},
        {"lesbian", color.New(color.FgHiRed).Add(color.BgHiBlack)},
        {"bi", color.New(color.FgHiMagenta).Add(color.BgHiBlue)},
        {"sunset", color.New(color.FgHiBlack).Add(color.BgHiYellow)},
        {"bg-sunset", color.New(color.BgHiYellow)},
        {"rainbow", color.New(color.FgHiBlack).Add(color.BgHiCyan)},
        {"bg-rainbow", color.New(color.BgHiCyan)},
    }
)


func colorNameWithColor(c string) string {
	style, _ := getStyle(c)
	return style.apply(c)
}

func init() {
	markdown.CurrentTheme = chromastyles.ParaisoDark
	ColorHelpMsg = strings.Join(func() []string {
		colors := make([]string, 0, len(Styles))
		for i := range Styles {
			colors = append(colors, Styles[i].apply(Styles[i].name))
		}
		return colors
	}(), ", ") + `

Make your own colors using hex (eg. color ` + colorNameWithColor("#A0FFFF") + ` or RGB values from 0-5 (eg. color ` + colorNameWithColor("530") + `). Generate gradients with hues (eg. color ` + colorNameWithColor("hue-0-360") + `). Set bg with "bg-" (eg. color ` + colorNameWithColor("bg-101") + `, color ` + colorNameWithColor("bg-hue-130-20") + `). Use multiple colors at once (eg. color ` + colorNameWithColor("rose #F5A9B8") + `). Remove bg with bg-off. There's also a few secret colors :)`
}

type Style struct {
	name  string
	apply func(string) string
}

func buildStyle(c *gchalk.Builder) func(string) string {
	return func(s string) string { return c.Paint(stripansi.Strip(s)) }
}
func buildStyleNoStrip(c *gchalk.Builder) func(string) string {
	return func(s string) string { return c.Paint(s) }
}

func makeFlag(colors []string) func(a string) string {
	flag := make([]*gchalk.Builder, len(colors))
	for i := range colors {
		flag[i] = TrueColor.WithHex(colors[i])
	}
	return func(a string) string {
		return applyMulticolor(flag, a)
	}
}

func rainbow(a string, bg bool) string {
	span := 360.0
	length := len(stripansi.Strip(a))
	if length < 16 {
		span = 22.5 * float64(length) // at least 45 degrees per letter
	}
	start := 360 * rand.Float64()
	//span /= 2
	return applyHueRange(start, start+span, a, bg)
}

func applyHueRange(start, end float64, a string, bg bool) string {
	if !bg { // if fg, strip all color (bg is applied after fg)
		a = stripansi.Strip(a)
	}
	buf := strings.Builder{}
	if !bg {
		runes := []rune(a)
		for i, r := range runes {
			h := start + (end-start)*float64(i)/float64(len(runes))
			buf.WriteString(TrueColor.WithRGB(hueRGB(h)).Paint(string(r)))
		}
	} else { // need to operate with ansi codes
		split := tokenizeAnsi(a)
		for i, s := range split {
			h := start + (end-start)*float64(i)/float64(len(split))
			buf.WriteString(TrueColor.WithBgRGB(hueRGB(h)).Paint(s))
		}
	}
	return buf.String()
}

func applyStyles(styles []*color.Color, a string) string {
    buf := strings.Builder{}
    colorOffset := rand.Intn(len(styles))
    for i, s := range tokenizeAnsi(a) {
        buf.WriteString(styles[(colorOffset+i)%len(styles)].Sprint(s))
    }
    return buf.String()
}

func applyMulticolor(colors []*gchalk.Builder, a string) string {
	a = stripansi.Strip(a)
	buf := strings.Builder{}
	colorOffset := rand.Intn(len(colors))
	for i, r := range []rune(a) {
		buf.WriteString(colors[(colorOffset+i)%len(colors)].Paint(string(r)))
	}
	return buf.String()
}

// splits runes and includes their color codes
func tokenizeAnsi(a string) []string {
	tokens := make([]string, 0, len(a)/3)
	buf := strings.Builder{}
	buildUntilM := false // m delineates end of ansi color code
	for _, r := range a {
		buf.WriteRune(r)
		if r == '\033' {
			buildUntilM = true
			continue
		}
		if buildUntilM {
			if r == 'm' {
				buildUntilM = false
			}
			continue
		}
		tokens = append(tokens, buf.String())
		buf.Reset()
	}
	if buf.Len() > 0 { // that last m could be needed
		tokens = append(tokens, buf.String())
	}

	for i := range tokens {
		if strings.HasPrefix(tokens[i], "\x1b[39m") {
			if i != len(tokens)-1 {
				tokens[i] = tokens[i][5:]
			} else {
				// move to earlier token
				tokens[i-1] += tokens[i]
				tokens = tokens[:len(tokens)-1]
			}
		}
	}
	return tokens
}

// h from 0 to 360
// https://www.desmos.com/calculator/wb91fw4nyj
func hueRGB(h float64) (r, g, b uint8) {
	pi := math.Pi
	h = math.Mod(h, 360) / 360.0
	r = uint8(math.Round(255.0 * (0.5 + 0.5*math.Sin(2*pi*h+pi/2))))
	g = uint8(math.Round(255.0 * (0.5 + 0.5*math.Sin(2*pi*h+pi/2+2*pi/3))))
	b = uint8(math.Round(255.0 * (0.5 + 0.5*math.Sin(2*pi*h+pi/2+4*pi/3))))
	//r, g, b, err := colorconv.HSVToRGB(math.Mod(h, 360), s, v)
	//if err != nil {
	//	return Chalk.WithRGB(0, 0, 0)
	//}
	return
}

// with r, g and b values from 0 to 5
func ansi256(r, g, b uint8) *gchalk.Builder {
	return Chalk.WithRGB(255/5*r, 255/5*g, 255/5*b)
	//return Chalk.WithRGB(uint8(math.Round(255*float64(r)/5)), uint8(math.Round(255*float64(g)/5)), uint8(math.Round(255*float64(b)/5)))
}

func bgAnsi256(r, g, b uint8) *gchalk.Builder {
	return Chalk.WithBgRGB(255/5*r, 255/5*g, 255/5*b)
}

// Applies color from name
func (u *User) changeColor(colorName string) error {
    if strings.Contains(colorName, "bg") {
        if names := strings.Fields(colorName); len(names) > 1 { // do we need to separate bg and fg colors?
            fgColors := make([]string, 0, len(names)-1)
            bgColors := make([]string, 0, len(names))
            for _, name := range names {
                if strings.HasPrefix(name, "bg") {
                    bgColors = append(bgColors, name)
                } else {
                    fgColors = append(fgColors, name)
                }
            }
            if len(fgColors) != 0 { // if no fg colors, carry on normally
                err := u.changeColor(strings.Join(fgColors, " "))
                if err != nil {
                    return err
                }
                return u.changeColor(strings.Join(bgColors, " "))
            }
        }
    }

    style, err := getStyle(colorName)
    if err != nil {
        return err
    }

    if strings.HasPrefix(colorName, "bg") {
        u.ColorBG = style.Name() // update bg color
    } else {
        u.Color = style.Name() // update fg color
    }

    u.Name, _ = applyColorToData(u.Name, u.Color, u.ColorBG)
    u.formatPrompt()
    return nil
}

	style, err := getStyle(colorName)
	if err != nil {
		return err
	}

	//changedBg := false
	if strings.HasPrefix(colorName, "bg") {
		//changedBg = true
		u.ColorBG = style.name // update bg color
	} else {
		u.Color = style.name // update fg color
	}

	u.Name, _ = applyColorToData(u.Name, u.Color, u.ColorBG)
	//styleFG := &Style{}
	//styleBG := &Style{}
	//if changedBg {
	//	styleFG, _ = getStyle(u.Color) // already checked for errors
	//	styleBG = style
	//} else {
	//	styleBG, err = getStyle(u.ColorBG) // already checked for errors
	//	styleFG = style
	//}
	//u.Name = styleBG.apply(styleFG.apply(u.Name))
	u.formatPrompt()
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
		return &Style{name, buildStyle(TrueColor.WithHex(name))}, nil
	}
	if strings.HasPrefix(name, "bg-#") {
		return &Style{name, buildStyleNoStrip(TrueColor.WithBgHex(strings.TrimPrefix(name, "bg-")))}, nil
	}
	bghue := strings.HasPrefix(name, "bg-hue-")
	if bghue || strings.HasPrefix(name, "hue-") {
		var hue1, hue2 float64
		var err error
		if bghue {
			_, err = fmt.Sscanf(name, "bg-hue-%f-%f", &hue1, &hue2)
		} else {
			_, err = fmt.Sscanf(name, "hue-%f-%f", &hue1, &hue2)
		}
		if err != nil {
			return nil, err
		}
		return &Style{name, func(a string) string { return applyHueRange(hue1, hue2, a, bghue) }}, nil
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
func getStyle(name string) (*color.Color, error) {
    name = strings.TrimSpace(name)
    if names := strings.Fields(name); len(names) > 1 {
        styleSlice := make([]*color.Color, len(names))
        newName := ""
        for i := range names {
            style := getStyle(names[i])
            if style == nil {
                return nil, errors.New("Unknown color: " + names[i])
            }
            styleSlice[i] = style
            newName += style.Name() + " "
        }

        return &color.Color{Name: newName[:len(newName)-1], ApplyFunc: func(a string) string {
            return applyStyles(styleSlice, a)
        }}, nil
    }

    for _, style := range Styles {
        if style.Name() == name {
            return style, nil
        }
    }

    for _, style := range SecretStyles {
        if style.Name() == name {
            return style, nil
        }
    }

    return nil, errors.New(`Which color? Choose from ` + colorNameWithColor("random") + `, ` + colorNameWithColor("bg-random") + `, ` + ColorHelpMsg)
}

// Return the msg string with the same colors as the reference string
func copyColor(msg string, ref string) string {
	stripedMsg := stripansi.Strip(msg)
	colorTokens := tokenizeAnsi(ref)
	ret := ""
	for i, c := range stripedMsg {
		token := colorTokens[i%len(colorTokens)]
		token = strings.ReplaceAll(token, "\033[39m", "") // Remove reset to default foreground and background
		tokenByte := []byte(strings.ReplaceAll(token, "\033[49m", ""))
		tokenByte[len(tokenByte)-1] = byte(c)
		ret += string(tokenByte)
	}
	ret += "\033[39m\033[49m"
	return ret
}
