package gover

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/restic/chunker"
)

func (snap Snapshot) Write(snapFilename string) {
	snapFolder := filepath.Join(WorkingDirectory, ".gover2", "snapshots")
	os.MkdirAll(snapFolder, 0777)
	snapFile := filepath.Join(snapFolder, snapFilename+".json")
	snap.WriteFile(snapFile)
}

func (snap Snapshot) WriteFile(snapshotPath string) {
	f, err := os.Create(snapshotPath)

	if err != nil {
		panic(fmt.Sprintf("Error: Could not create snapshot file %s", snapshotPath))
	}
	myEncoder := json.NewEncoder(f)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(snap)
	f.Close()
}

func ReadSnapshot(snapId string) Snapshot {
	snapshotPath := filepath.Join(".gover2", "snapshots", snapId+".json")

	if VerboseMode {
		fmt.Printf("Reading %s\n", snapshotPath)
	}

	return ReadSnapshotFile(snapId)
}

// Read a snapshot given a file path
func ReadSnapshotFile(snapshotPath string) Snapshot {
	var snap Snapshot
	f, err := os.Open(snapshotPath)

	// ChunkPackIds map[string]string
	// FileChunkIds map[string][]string
	// FileModTimes map[string]string

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		snap := Snapshot{}
		snap.ChunkPackIds = make(map[string]string)
		snap.FileChunkIds = make(map[string][]string)
		snap.FileModTimes = make(map[string]string)
		return snap
	}

	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&snap); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", snapshotPath)
		Check(err)
	}

	f.Close()
	return snap
}

func WriteHead(snapshotPath string) {
	headPath := filepath.Join(".gover2", "head.json")
	f, err := os.Create(headPath)

	if err != nil {
		panic(fmt.Sprintf("Error: Could not create head file %s", headPath))
	}

	myEncoder := json.NewEncoder(f)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(snapshotPath)
	f.Close()
}

// Read a snapshot given a file path
func ReadHead() Snapshot {
	headPath := filepath.Join(".gover2", "head.json")
	f, err := os.Open(headPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		snap := Snapshot{}
		snap.ChunkPackIds = make(map[string]string)
		snap.FileChunkIds = make(map[string][]string)
		snap.FileModTimes = make(map[string]string)
		return snap
	}

	snapshotId := ""
	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&snapshotId); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", headPath)
		Check(err)
	}

	f.Close()

	snapshotPath := filepath.Join(".gover2", "snapshots", snapshotId+".json")
	return ReadSnapshotFile(snapshotPath)
}

func CreatePackFile(packId string) (*os.File, error) {
	goverDir := filepath.Join(WorkingDirectory, ".gover2")
	packFolderPath := path.Join(goverDir, "packs", packId[0:2])
	os.MkdirAll(packFolderPath, 0777)
	packPath := path.Join(packFolderPath, packId+".zip")

	if VerboseMode {
		fmt.Printf("Creating pack: %s\n", packId[0:16])
	}

	// TODO: only create pack file if we need to save stuff - set to nil initially
	packFile, err := os.Create(packPath)

	if err != nil {
		if VerboseMode {
			fmt.Printf("Error creating pack file %s", packPath)
		}

		return nil, err
	}

	return packFile, nil
}

func WriteChunkToPack(zipWriter *zip.Writer, chunkId string, chunk chunker.Chunk) error {
	var header zip.FileHeader
	header.Name = chunkId
	header.Method = CompressionLevel

	writer, err := zipWriter.CreateHeader(&header)

	if err != nil {
		if VerboseMode {
			fmt.Printf("Error creating zip header\n")
		}

		return err
	}

	if _, err := writer.Write(chunk.Data); err != nil {
		if VerboseMode {
			fmt.Printf("Error writing chunk %s to zip file\n", chunkId)
		}

		return err
	}

	return nil
}

func ExtractChunkFromPack(outFile *os.File, chunkId string, packId string) error {
	goverDir := filepath.Join(WorkingDirectory, ".gover2")
	packFolderPath := path.Join(goverDir, "packs", packId[0:2])
	packPath := path.Join(packFolderPath, packId+".zip")
	packFile, err := zip.OpenReader(packPath)

	if err != nil {
		if VerboseMode {
			fmt.Printf("Error extracting pack %s[%s]\n", packId, chunkId)
		}
		return err
	}

	defer packFile.Close()
	return ExtractChunkFromZipFile(outFile, packFile, chunkId)
}

func ExtractChunkFromZipFile(outFile *os.File, packFile *zip.ReadCloser, chunkId string) error {
	for _, f := range packFile.File {

		if f.Name == chunkId {
			// fmt.Printf("Contents of %s:\n", f.Name)
			chunkFile, err := f.Open()

			if err != nil {
				if VerboseMode {
					fmt.Printf("Error opening chunk %s\n", chunkId)
				}

				return err
			}

			_, err = io.Copy(outFile, chunkFile)

			if err != nil {
				if VerboseMode {
					fmt.Printf("Error reading chunk %s\n", chunkId)
				}

				return err
			}

			chunkFile.Close()
		}
	}

	return nil
}
