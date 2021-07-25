package build

import "embed"

//go:embed embedded/.deploy-pgp embedded/README.md embedded/run-wrapper.sh
var embedded embed.FS

// Asset reads the embedded file and returns it.
func Asset(file string) ([]byte, error) {
	return embedded.ReadFile(file)
}
