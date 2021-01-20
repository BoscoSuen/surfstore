package surfstore

type State int

const (
    Unchanged State = iota
    Modified
    New
    Deleted
)

type FileInfo struct {
    FileMetaData   FileMetaData
    Status         State
}
