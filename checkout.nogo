package gover

import (
	"fmt"
	"os"
	"path/filepath"
)

func CheckoutSnaphot(snapshotNum int, outputFolder string) {
	if len(outputFolder) == 0 {
		outputFolder = fmt.Sprintf("snapshot%04d", snapshotNum)
	}

	fmt.Printf("Checking out %s\n", snapshotNum)

	snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
	snapshotPaths, err := filepath.Glob(snapshotGlob)
	Check(err)

	snapshotPath := snapshotPaths[snapshotNum-1]
	fmt.Printf("Reading %s\n", snapshotPath)
	snap := ReadSnapshotFile(snapshotPath)

	os.Mkdir(outputFolder, 0777)

	for _, file := range snap.Files {
		fileDir := filepath.Dir(file)
		outDir := outputFolder

		if fileDir != "." {
			outDir = filepath.Join(outputFolder, fileDir)
			fmt.Printf("Creating folder %s\n", outDir)
			os.MkdirAll(outDir, 0777)
		}

		outFile := filepath.Join(outputFolder, file)
		storedFile := snap.StoredFiles[file]
		fmt.Printf("Restoring %s to %s\n", storedFile, outFile)
		CopyFile(storedFile, outFile)
	}
}
