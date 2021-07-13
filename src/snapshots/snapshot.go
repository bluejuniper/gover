package snapshots

type Snapshot struct {
	Message  string
	Time     string
	Files    []string
	PacksIds map[string][]string
	ChunkIds map[string][]string
	ModTimes map[string]string
	Sizes    map[string]int64
}
