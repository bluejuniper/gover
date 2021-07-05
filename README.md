# Gover
Go utility for binary version control

This is a much simplified version of [dupver](https://github.com/akbarnes/dupver/) that uses a local repository and file-level deduplication. There are only four commands `commit`, `status`, `log`, and `checkout`. 

To build:
``` go
go mod init goutils
go get github.com/bmatcuk/doublestar/v4
go install gover.go
```

### Commit
The `-msg` or `-m` message flag is optional, as is the `-commit` or `-ci` flag as commiting is the default action
`gover -commit -msg 'a message' file1 file2 file3`
`gover -ci -msg 'a message' file1 file2 file3`
`gover -m  'a message 'file1 file2 file3`
`gover file1 file2 file3`

### Log
This takes the optional `-json` or `-j` argument to output json for use with object shells. To list all the snapshots:
``` bash
gover -log
gover -json -log
gover -j -log
gover -l
```

To list the files in a particular snapshot:
`gover -log snapshot_time`
`gover -log -json snapshot_time`

## Checkout
This takes an optional argument to specify an output folder. To checkout a snapshot:
`gover -checkout snapshot_time`
`gover -co snapshot_time`
`gover -out output_folder -co snapshot_time`
`gover -o output_folder -co snapshot_time`
