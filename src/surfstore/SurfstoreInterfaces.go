package surfstore

type Block struct {
    BlockData []byte
    BlockSize int
}

type FileMetaData struct {
    Filename      string
    Version       int
    BlockHashList []string
}

type Surfstore interface {
    MetaStoreInterface
    BlockStoreInterface
}

type MetaStoreInterface interface {
    // Retrieves the server's FileInfoMap
    GetFileInfoMap(_ignore *bool, serverFileInfoMap *map[string]FileMetaData) error

    // Update a file's fileinfo entry
    UpdateFile(fileMetaData *FileMetaData, latestVersion *int) (err error)
}

type BlockStoreInterface interface {

    // Get a block based on its hash
    GetBlock(blockHash string, block *Block) error

    // Put a block
    PutBlock(block Block, succ *bool) error

    // Check if certain blocks are alredy present on the server
    HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error
}
