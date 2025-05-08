// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
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
func calculateHash(path string, baseImage string) string {
	hasher := sha256.New()

	// Walk through the directory recursively
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it's a file, get its size
		if !info.IsDir() {
			fileSize := info.Size()
			// covert int64 to byte[]
			fileSizeBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(fileSizeBytes, uint64(fileSize))
			hasher.Write(fileSizeBytes)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", path, err)
		return ""
	}

	hasher.Write([]byte(baseImage))

	// Get the final hash sum
	hashSum := hasher.Sum(nil)

	// Convert the hash sum to a hexadecimal string
	hashCode := fmt.Sprintf("%x", hashSum)

	return hashCode
}
