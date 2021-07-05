package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"strings"
	"strconv"
	"path/filepath"
	"encoding/json"
	"github.com/bmatcuk/doublestar/v4"
)

const NumChars = 40

type Snapshot struct {
	Message       string
	Time          string
	Files	      []string
	StoredFiles	  map[string]string
	ModTimes	  map[string]string
}

func check(e error) {
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

func WriteHead(snapshotPath string) {
	headPath := filepath.Join(".gover", "head.json")
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
	headPath := filepath.Join(".gover", "head.json")
	f, err := os.Open(headPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		return Snapshot{Files: []string{}, StoredFiles: make(map[string]string), ModTimes: make(map[string]string)}
	}

	snapshotId := ""
	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&snapshotId); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", headPath)
		check(err)
	}

	f.Close()

	snapshotPath := filepath.Join(".gover", "snapshots", snapshotId + ".json")
	return ReadSnapshotFile(snapshotPath)
}

func (snap Snapshot) Write(snapshotPath string) {
	f, err := os.Create(snapshotPath)

	if err != nil {
		panic(fmt.Sprintf("Error: Could not create snapshot file %s", snapshotPath))
	}
	myEncoder := json.NewEncoder(f)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(snap)
	f.Close()
}


func CommitSnapshot(message string, filters []string) {
	// Optional timestamp
	t := time.Now()
	ts := t.Format("2006-01-02T15-04-05")
	snap := Snapshot{Time: ts}
	snap.Files = []string{}
	snap.StoredFiles = make(map[string]string)
	snap.ModTimes = make(map[string]string)
	snap.Message = message

	// workingDirectory, err := os.Getwd()
	// check(err)
	workingDirectory := "."
	head := ReadHead()

	goverDir := filepath.Join(workingDirectory, ".gover", "**")

	var VersionFile = func(fileName string, info os.FileInfo, err error) error {
		fileName = strings.TrimSuffix(fileName, "\n")

		if info.IsDir() {
			return nil
		}

		matched, err := doublestar.PathMatch(goverDir, fileName)

		if matched {
			if VerboseMode {
				fmt.Printf("Skipping file %s in .gover\n", fileName)
			}

			return nil
		}

		for _, pattern := range filters {
			matched, err := doublestar.PathMatch(pattern, fileName)

			check(err)
			if matched {
				if VerboseMode {
					fmt.Printf("Skipping file %s which matches with %s\n", fileName, pattern)
				}

				return nil
			}
		}

		ext := filepath.Ext(fileName)
		hash, hashErr := HashFile(fileName, NumChars)

		if hashErr != nil {
			return hashErr
		}

		verFolder := filepath.Join(".gover", "data", hash[0:2]) 
		verFile := filepath.Join(verFolder, hash + ext)

		props, err := os.Stat(fileName)

		if err != nil {
			if VerboseMode {
				fmt.Printf("Skipping unreadable file %s\n", fileName)
			}

			return nil
		}

		modTime := props.ModTime().Format("2006-01-02T15-04-05")



		snap.Files = append(snap.Files, fileName)
		snap.StoredFiles[fileName] = verFile
		snap.ModTimes[fileName] = modTime

		os.MkdirAll(verFolder, 0777)

		if headModTime, ok := head.ModTimes[fileName]; ok && modTime == headModTime {
			if VerboseMode {
				fmt.Printf("Skipping %s\n", fileName)
			}
		} else {
			CopyFile(fileName, verFile)

			if !JsonMode {
				if VerboseMode {
						fmt.Printf("%s -> %s\n", fileName, verFile)
				} else {
					fmt.Println(fileName)
				}
			}
		}

		return nil
	}

	// fmt.Printf("No changes detected in %s for commit %s\n", workDir, snapshot.ID)
	filepath.Walk(workingDirectory, VersionFile)

	if JsonMode {
		PrintJson(snap)
	}

	snapFolder := filepath.Join(".gover", "snapshots")
	os.MkdirAll(snapFolder, 0777)
	snapFile := filepath.Join(snapFolder, ts + ".json")
	snap.Write(snapFile)
	WriteHead(ts)
}

func DiffSnapshot(snapId string, filters []string) {
	var snap Snapshot

	if len(snapId) > 0 {
		snap = ReadSnapshot(snapId)
	} else {
		snap = ReadHead()
	}

	status := make(map[string]string)

	for _, fileName := range snap.Files {
		status[fileName] = "-"
	}

	// workingDirectory, err := os.Getwd()
	// check(err)
	workingDirectory := "."
	head := ReadHead()

	goverDir := filepath.Join(workingDirectory, ".gover", "**")

	var DiffFile = func(fileName string, info os.FileInfo, err error) error {
		fileName = strings.TrimSuffix(fileName, "\n")

		if info.IsDir() {
			return nil
		}

		matched, err := doublestar.PathMatch(goverDir, fileName)

		if matched {
			if VerboseMode {
				fmt.Printf("Skipping file %s in .gover\n", fileName)
			}

			return nil
		}

		for _, pattern := range filters {
			matched, err := doublestar.PathMatch(pattern, fileName)

			check(err)
			if matched {
				if VerboseMode {
					fmt.Printf("Skipping file %s which matches with %s\n", fileName, pattern)
				}

				return nil
			}
		}

		ext := filepath.Ext(fileName)
		hash, hashErr := HashFile(fileName, NumChars)

		if hashErr != nil {
			return hashErr
		}

		verFolder := filepath.Join(".gover", "data", hash[0:2]) 
		verFile := filepath.Join(verFolder, hash + ext)

		props, err := os.Stat(fileName)

		if err != nil {
			if VerboseMode {
				fmt.Printf("Skipping unreadable file %s\n", fileName)
			}

			return nil
		}

		modTime := props.ModTime().Format("2006-01-02T15-04-05")



		snap.Files = append(snap.Files, fileName)
		snap.StoredFiles[fileName] = verFile
		snap.ModTimes[fileName] = modTime

		os.MkdirAll(verFolder, 0777)

		if headModTime, ok := head.ModTimes[fileName]; ok {
			if modTime == headModTime {
					status[fileName] = "="
			} else {
				status[fileName] = "M"
			}
		} else {
			status[fileName] = "+"
		}

		return nil
	}

	// fmt.Printf("No changes detected in %s for commit %s\n", workDir, snapshot.ID)
	filepath.Walk(workingDirectory, DiffFile)

	if JsonMode {

	} else {
		for fileName, fileStatus := range status {
			if fileStatus == "=" && !VerboseMode {
				continue
			}
	
			fmt.Printf("%s %s\n", fileStatus, fileName)
		}
	}
}

func CheckoutSnaphot(snapshotNum int, outputFolder string) {
	if len(outputFolder) == 0 {
		outputFolder = fmt.Sprintf("snapshot%04d", snapshotNum)
	}

	fmt.Printf("Checking out %s\n", snapshotNum)

	snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
	snapshotPaths, err := filepath.Glob(snapshotGlob)
	check(err)

	snapshotPath := snapshotPaths[snapshotNum - 1]
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

func LogSingleSnapshot(snapshotNum int) {
	snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
	snapshotPaths, err := filepath.Glob(snapshotGlob)
	check(err)

	snapshotPath := snapshotPaths[snapshotNum - 1]

	snap := ReadSnapshotFile(snapshotPath)

	if JsonMode {
		type SnapshotFile struct {
			File string
			StoredFile string
		}

		snapFiles := []SnapshotFile{}

		for _, file := range snap.Files {
			snapFile := SnapshotFile{File: file, StoredFile:snap.StoredFiles[file]}
			snapFiles = append(snapFiles, snapFile)
		}

		PrintJson(snapFiles)
	} else {
		for _, file := range snap.Files {
			fmt.Println(file)
		}		
	}	
}

func LogAllSnapshots() {
	if JsonMode {
		type Snap struct {
			Time string
			Message string
		}

		snaps := []Snap{}
		
		snapshotGlob := filepath.Join(".gover","snapshots","*.json")
		snapshotPaths, err := filepath.Glob(snapshotGlob)
		check(err)

		for _, snapshotPath := range snapshotPaths {
			snapshot := ReadSnapshotFile(snapshotPath)
			snap := Snap{Time: snapshot.Time, Message: snapshot.Message}
			snaps = append(snaps, snap)
		}

		PrintJson(snaps)
	} else {
		snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
		snapshotPaths, err := filepath.Glob(snapshotGlob)
		check(err)

		for i, snapshotPath := range snapshotPaths {
			snap := ReadSnapshotFile(snapshotPath)
			// Time: 2021/05/08 08:57:46
			// Message: specify workdir path explicitly
			fmt.Printf("%3d) Time: %s\n", i + 1, snap.Time)

			if len(snap.Message) > 0 {
				fmt.Printf("Message: %s\n\n", snap.Message)
			}
		}
	}
}

func ReadSnapshot(snapId string) Snapshot {
	snapshotPath := filepath.Join(".gover", "snapshots", snapId + ".json")

	if VerboseMode {
		fmt.Printf("Reading %s\n", snapshotPath)
	}

	return ReadSnapshotFile(snapId)
}

// Read a snapshot given a file path
func ReadSnapshotFile(snapshotPath string) Snapshot {
	var mySnapshot Snapshot
	f, err := os.Open(snapshotPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		return Snapshot{Files: []string{}, StoredFiles: make(map[string]string), ModTimes: make(map[string]string)}
	}

	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&mySnapshot); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", snapshotPath)
		check(err)
	}

	f.Close()
	return mySnapshot
}

// Print an object as JSON to stdout
func PrintJson(a interface{}) {
	myEncoder := json.NewEncoder(os.Stdout)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(a)
}

func ReadFilters() []string {
	filterPath := ".gover_ignore.json"
	var filters []string
	f, err := os.Open(filterPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		return []string{}
	}

	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&filters); err != nil {
		panic(fmt.Sprintf("Error:could not decode filter file %s", filterPath))
	}

	f.Close()
	return filters
}

var JsonMode bool
var Message string
var OutputFolder string
var VerboseMode bool

func init() {
	flag.BoolVar(&JsonMode, "json", false, "print json")
	flag.BoolVar(&JsonMode, "j", false, "print json")
	flag.BoolVar(&VerboseMode, "verbose", false, "verbose")
	flag.BoolVar(&VerboseMode, "v", false, "verbose")	
}

// type Commit struct {
// 	ID        string
// 	Branch    string
// 	Message   string
// 	Time      string
// 	ParentIDs []string
// 	Files     []fileInfo
// 	ChunkIDs  []string
// }

func main() {
	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	logCmd := flag.NewFlagSet("log", flag.ExitOnError)
	checkoutCmd := flag.NewFlagSet("checkout", flag.ExitOnError)
	
	flag.Parse()	
	
	if len(os.Args) < 2 {
        fmt.Println("Expected subcommand")
        os.Exit(1)
	}
	
	cmd := os.Args[1] 

	if cmd == "commit" || cmd == "ci" {
		commitCmd.Parse(os.Args[2:])
		filters := ReadFilters()

		if commitCmd.NArg() >= 1 {
			Message = commitCmd.Arg(0)
		}

		CommitSnapshot(Message, filters)
	} else if cmd == "status" || cmd == "st" {
		statusCmd.Parse(os.Args[2:])
		filters := ReadFilters()

		if statusCmd.NArg() >= 1 {
			DiffSnapshot(statusCmd.Arg(0), filters)
		} else {
			DiffSnapshot("", filters)
		}		
	} else if cmd == "log" {
		logCmd.Parse(os.Args[2:])

		if logCmd.NArg() >= 1 {
			snapshotNum, _ := strconv.Atoi(logCmd.Arg(0))
			LogSingleSnapshot(snapshotNum)
		} else {
			LogAllSnapshots()
		}
	} else if cmd == "checkout" || cmd == "co" {
		checkoutCmd.StringVar(&OutputFolder, "out", "", "output folder")
		checkoutCmd.StringVar(&OutputFolder, "o", "", "output folder")	
		checkoutCmd.Parse(os.Args[2:])
		snapshotNum, _ := strconv.Atoi(checkoutCmd.Arg(0))
		CheckoutSnaphot(snapshotNum, OutputFolder)
	} else {
        fmt.Println("Unknown subcommand")
        os.Exit(1)
	}
}

