package updates

// Binaries define all binaries used by githooks.
type Binaries struct {
	Cli    string   // The Githooks cli binary.
	Others []string // All other binaries except the cli binary.
	All    []string // All binaries.

	BinDir string // Directory where all binaries reside.
}
