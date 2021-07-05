package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"encoding/json"
	"crypto/sha256" 
)

const NumChars = 40

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

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

// Print an object as JSON to stdout
func PrintJson(a interface{}) {
	myEncoder := json.NewEncoder(os.Stdout)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(a)
}
