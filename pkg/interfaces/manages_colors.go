package interfaces

import (
	"devzat/pkg/colors"
	"github.com/jwalton/gchalk"
)

type managesColors interface {
	ApplyRainbow(rainbow []*gchalk.Builder, a string) string
	Ansi256(r, g, b uint8) *gchalk.Builder
	BgAnsi256(r, g, b uint8) *gchalk.Builder
	ApplyColorToData(data string, color string, colorBG string) (string, error)
	GetRandomColor(name string) *colors.Style
	GetNamedColor(name string) *colors.Style
	GetCustomColor(name string) (*colors.Style, error)
	GetStyle(name string) (*colors.Style, error)
	GetStyleNames() []string
	Colors() colors.Colors
}
