package gover

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func CheckoutSnaphot(snapshotNum int, outputFolder string) {
	goverDir := filepath.Join(WorkingDirectory, ".gover2")

	if len(outputFolder) == 0 {
		outputFolder = fmt.Sprintf("snapshot%04d", snapshotNum)
	}

	fmt.Printf("Checking out %s\n", snapshotNum)

	snapshotGlob := filepath.Join(goverDir, "snapshots", "*.json")
	snapshotPaths, err := filepath.Glob(snapshotGlob)
	Check(err)

	snapshotPath := snapshotPaths[snapshotNum-1]
	fmt.Printf("Reading %s\n", snapshotPath)
	snap := ReadSnapshotFile(snapshotPath)

	os.Mkdir(outputFolder, 0777)

	for file, _ := range snap.FileModTimes {
		fileDir := filepath.Dir(file)
		outDir := outputFolder

		if fileDir != "." {
			outDir = filepath.Join(outputFolder, fileDir)
			fmt.Printf("Creating folder %s\n", outDir)
			os.MkdirAll(outDir, 0777)
		}

		outPath := filepath.Join(outputFolder, file)
		outFile, err := os.Create(outPath)
		Check(err)
		defer outFile.Close()

		for _, chunkId := range snap.FileChunkIds[file] {
			packFolderPath := path.Join(goverDir, "packs", packId[0:2])
			packPath := path.Join(packFolderPath, packId+".zip")
			packFile, err := zip.OpenReader(packPath)
			Check(err)
			defer packFile.Close()

		}

		fmt.Printf("Restored %s to %s\n", file, outPath)
	}
}
