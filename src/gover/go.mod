module github.com/akbarnes/src/gover

go 1.16

replace github.com/akbarnes/gover/src/snapshots => ../snapshots

replace github.com/akbarnes/gover/src/util => ../util

require (
	github.com/akbarnes/gover/src/options v0.0.0-00010101000000-000000000000
	github.com/akbarnes/gover/src/snapshots v0.0.0-00010101000000-000000000000
	github.com/akbarnes/gover/src/util v0.0.0-20210705125008-9c9527b91849
	github.com/bmatcuk/doublestar/v4 v4.0.2
	github.com/restic/chunker v0.4.0 // indirect
)

replace github.com/akbarnes/gover/src/options => ../options
