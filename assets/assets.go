package assets

import "embed"

//go:embed *
var content embed.FS

func ReadFile(name string) ([]byte, error) {
	return content.ReadFile(name)
}
