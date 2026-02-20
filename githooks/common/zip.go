package common

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ExtractZip will decompress a zip archive, moving all files and folders
// within the zip file in `src` to an output directory `dest`.
// Returns all file paths in the zip file.
func ExtractZip(
	zipFile io.ReaderAt,
	zipFileSize int64,
	destDir string,
) (paths []string, err error) {
	r, err := zip.NewReader(zipFile, zipFileSize)
	if err != nil {
		return paths, err
	}

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := path.Join(destDir, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			err = ErrorF("Unzip: illegal file path '%s'.", fpath)

			return paths, err
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return paths, err
			}

			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return paths, err
		}

		var outFile *os.File
		outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return paths, err
		}

		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			return paths, err
		}

		_, err = io.Copy(outFile, rc)
		// Close the file without defer to close before next iteration of loop
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return paths, err
		}

		paths = append(paths, fpath)
	}

	return paths, nil
}
