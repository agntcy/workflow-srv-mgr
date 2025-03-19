package python

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// CalculateHash calculates a hash code for the given path by iterating over all files and folders
// recursively and using the size of each file.
func calculateHash(path string) string {
	hasher := sha256.New()

	// Walk through the directory recursively
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it's a file, get its last modified time
		if !info.IsDir() {
			modTime := info.Size()
			// Convert the modification time to bytes and add it to the hash
			binary.Write(hasher, binary.LittleEndian, modTime)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", path, err)
		return ""
	}

	// Get the final hash sum
	hashSum := hasher.Sum(nil)

	// Convert the hash sum to a hexadecimal string
	hashCode := fmt.Sprintf("%x", hashSum)

	return hashCode
}
