package download

import (
	cm "github.com/gabyx/githooks/githooks/common"
	archiver "github.com/mholt/archiver/v3"
)

// Extract extracts a file int dir.
// The extension guides the type of extraction.
func Extract(file string, extension string, dir string) (err error) {
	err = archiver.Unarchive(file, dir)
	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF(
			"Could not extract downloaded file '%s' into '%s'.", file, dir))
	}

	return
}
