package util

import (
	_ "embed"
	"io/ioutil"
)

func GetAsciiArt() string {
	b, _ := ioutil.ReadFile("art.txt")
	if b == nil {
		return "sowwy, no art was found, please slap your developer and tell em to add an art.txt file"
	}
	return string(b)
}
