package gover

type Snapshot struct {
	Message     string
	Time        string
	Files       []string
	StoredFiles map[string]string
	ModTimes    map[string]string
}
