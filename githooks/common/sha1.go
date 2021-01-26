package common

import (
	"crypto/sha1"
	"encoding/hex"
	strs "gabyx/githooks/strings"
	"io"
	"os"
)

// GetSHA1HashFile gets the SHA1 hash of a file.
// It properly mimics `git hash-file`.
func GetSHA1HashFile(path string) (sha string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return
	}

	stat.Size()

	// Open a new SHA1 hash interface to write to
	hash := sha1.New()

	// Mimic a Git SHA1 hash.
	_, err = strs.FmtW(hash, "blob %v\u0000", stat.Size())
	if err != nil {
		return "", err
	}

	// Copy the file in the hash interface and check for any error
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	sha = hex.EncodeToString(hash.Sum(nil))

	return
}

// GetSHA1Hash gets the SHA1 hash of a string.
func GetSHA1Hash(reader io.Reader) (string, error) {
	h := sha1.New()

	if _, err := io.Copy(h, reader); err != nil {
		return "", nil
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
