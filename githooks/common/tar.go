package common

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"runtime"
)

// ExtractTarGz extracts `.tar.gz` streams to `baseDir` which must not exist.
// Overwrites everything.
func ExtractTarGz(gzipStream io.Reader, baseDir string) (paths []string, err error) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return
	}

	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header

	err = os.MkdirAll(baseDir, DefaultFileModeDirectory)
	if err != nil {
		return
	}

	for {
		header, err = tarReader.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		outPath := path.Join(baseDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:

			err = os.MkdirAll(outPath, DefaultFileModeDirectory)
			if err != nil {
				return
			}

		case tar.TypeReg:

			var file *os.File
			file, err = os.Create(outPath)
			if err != nil {
				return
			}
			defer func() { _ = file.Close() }()

			if _, err = io.Copy(file, tarReader); err != nil {
				err = CombineErrors(ErrorF("Copy of data to '%s' failed", outPath), err)

				return
			}

			if runtime.GOOS == WindowsOsName {
				_ = file.Close()
				if err = Chmod(outPath, header.FileInfo().Mode()); err != nil {
					return
				}
			} else if err = file.Chmod(header.FileInfo().Mode()); err != nil {
				return
			}

			paths = append(paths, outPath)

		default:
			err = ErrorF("Tar extracting: unknown type: '%v' in '%v'",
				header.Typeflag,
				header.Name)

			return
		}
	}

	return paths, nil
}
