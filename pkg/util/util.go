package util

import (
	"bytes"
	"devzat/pkg/interfaces"
	_ "embed"
	"fmt"
	"io/ioutil"
	"text/tabwriter"
)

func GetAsciiArt() string {
	b, _ := ioutil.ReadFile("art.txt")
	if b == nil {
		return "sowwy, no art was found, please slap your developer and tell em to add an art.txt file"
	}
	return string(b)
}

func AutoGenCommands(cmds []interfaces.Command) string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)

	for _, c := range cmds {
		formatted := fmt.Sprintf("   %s\t%s\t_%s_  \n", c.Name, c.ArgsInfo, c.Info)
		_, _ = w.Write([]byte(formatted))
	}

	_ = w.Flush()

	return b.String()
}
