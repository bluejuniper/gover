package gover

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

const NumChars = 40

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// TODO: merge these two togther to save one file read
func HashFile(FileName string, NumChars int) (string, error) {
	var data []byte
	var err error

	data, err = ioutil.ReadFile(FileName)

	if err != nil {
		return "", err
	}

	sum := fmt.Sprintf("%x", sha256.Sum256(data))

	if len(sum) < NumChars || NumChars < 0 {
		NumChars = len(sum)
	}

	return sum[0:NumChars], nil
}

// Copy the source file to a destination file. Any existing file
// will be overwritten and will not copy file attributes.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func ExcludedFile(fileName string, info os.FileInfo, filters []string) bool {
	// goverDir := filepath.Join(WorkingDirectory, ".gover2")
	goverDir := ".gover2"
	goverPattern := filepath.Join(goverDir, "**")

	if info.IsDir() {
		return true
	}

	matched, err := doublestar.PathMatch(goverPattern, fileName)

	if err != nil && VerboseMode {
		fmt.Printf("Error matching %s\n", goverDir)
	}

	if matched {
		if VerboseMode {
			fmt.Printf("Skipping file %s in .gover2\n", fileName)
		}

		return true
	}

	for _, pattern := range filters {
		matched, err := doublestar.PathMatch(pattern, fileName)

		if err != nil && VerboseMode {
			fmt.Printf("Error matching %s\n", goverDir)
		}

		if matched {
			if VerboseMode {
				fmt.Printf("Skipping file %s which matches with %s\n", fileName, pattern)
			}

			return true
		}
	}

	return false
}
