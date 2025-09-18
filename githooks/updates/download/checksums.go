package download

import (
	"bytes"
	"io"
	"os"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
)

// verifyChecksumSignature verifies checksums with the signature and the public key, and returns
// the checksums content.
func verifyChecksumSignature(checksums Checksums, publicPGP string) ([]byte, error) {
	checksumFile, err := GetFile(checksums.File.URL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = checksumFile.Body.Close() }()

	checksumFileSignature, err := GetFile(checksums.FileSignature.URL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = checksumFileSignature.Body.Close() }()

	// Read the checksumFile into memory
	checksumBytes, err := io.ReadAll(checksumFile.Body)
	if err != nil {
		return nil, err
	}

	err = cm.VerifyFile(bytes.NewReader(checksumBytes), checksumFileSignature.Body, publicPGP)
	if err != nil {
		return nil, err
	}

	return checksumBytes, nil
}

// checkChecksum checks if the checksum of file matches the checksum.
func checkChecksum(filePath string, checksumData []byte) (err error) {
	var file *os.File

	file, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	hash, err := cm.GetSHA256Hash(file)
	if err != nil {
		return err
	}

	if !strings.Contains(string(checksumData), hash) {
		return cm.ErrorF("Could not find checksum '%s' in checksum data.", filePath)
	}

	return nil
}
