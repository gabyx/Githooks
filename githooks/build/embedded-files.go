package build

import "embed"

//go:embed embedded/.deploy-pgp embedded/README.md embedded/run-wrapper.sh
var embedded embed.FS

func Asset(file string) ([]byte, error) {
	return embedded.ReadFile(file)
}
