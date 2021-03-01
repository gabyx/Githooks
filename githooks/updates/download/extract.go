package download

import (
	cm "gabyx/githooks/common"
	"os"
)

// Extract extracts a file int dir.
// The extension guides the type of extraction.
func Extract(file string, extension string, dir string) error {
	// Extract the file.
	f, err := os.Open(file)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Could not open file '%s'.", file))
	}
	defer f.Close()

	switch extension {
	case ".tar.gz":

		_, err = cm.ExtractTarGz(f, dir)

	case ".zip":

		var fi os.FileInfo
		fi, err = f.Stat()
		if err != nil {
			return cm.CombineErrors(err, cm.ErrorF("Could not get stats for file '%s'.", file))
		}
		_, err = cm.ExtractZip(f, fi.Size(), dir)

	default:
		cm.PanicF("Extraction not implemented for extension %s'", extension)
	}

	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF(
			"Could not extract downloaded file '%s' into '%s'.", file, dir))
	}

	return nil
}
