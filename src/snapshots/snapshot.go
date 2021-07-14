package snapshots

const PACK_ID_LEN int = 64

type Snapshot struct {
	Message      string
	Time         string
	ChunkPackIds map[string]string
	FileChunkIds map[string][]string
	FileModTimes map[string]string
}
