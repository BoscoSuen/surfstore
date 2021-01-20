package surfstore

import (
    "log"
    "net"
    "net/http"
    "net/rpc"
    "sync"
)

type Server struct {
    BlockStore BlockStoreInterface
    MetaStore  MetaStoreInterface
    Mutex      *sync.RWMutex
}

func (s *Server) GetFileInfoMap(succ *bool, serverFileInfoMap *map[string]FileMetaData) error {
    s.Mutex.RLock()
    defer func() {
        s.Mutex.RUnlock()
    }()
    err := s.MetaStore.GetFileInfoMap(succ, serverFileInfoMap)
    if err != nil {
        log.Println("GetFileInfoMap Error: ", err)
    }
    return err
}

func (s *Server) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
    s.Mutex.Lock()
    defer func() {
        s.Mutex.Unlock()
    }()
    err := s.MetaStore.UpdateFile(fileMetaData, latestVersion)
    if err != nil {
        log.Println("UpdateFile Error: ", err)
    }
    return err
}

func (s *Server) GetBlock(blockHash string, blockData *Block) error {
    s.Mutex.RLock()
    defer func() {
        s.Mutex.RUnlock()
    }()
    err := s.BlockStore.GetBlock(blockHash, blockData)
    if err != nil {
        log.Println("GetBlock Error: ", err)
    }
    return err
}

func (s *Server) PutBlock(blockData Block, succ *bool) error {
    s.Mutex.Lock()
    defer func() {
        s.Mutex.Unlock()
    }()
    err := s.BlockStore.PutBlock(blockData, succ)
    if err != nil {
        log.Println("PutBlock Error: ", err)
    }
    return err
}

func (s *Server) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
    s.Mutex.Lock()
    defer func() {
        s.Mutex.Unlock()
    }()
    err := s.BlockStore.HasBlocks(blockHashesIn, blockHashesOut)
    if err != nil {
        log.Println("HasBlocks Error: ", err)
    }
    return err
}   

// This line guarantees all method for surfstore are implemented
var _ Surfstore = new(Server)

func NewSurfstoreServer() Server {
    blockStore := BlockStore{BlockMap: map[string]Block{}}
    metaStore := MetaStore{FileMetaMap: map[string]FileMetaData{}}
    mutex := &sync.RWMutex{}

    return Server{
        BlockStore: &blockStore,
        MetaStore:  &metaStore,
        Mutex: mutex,
    }
}

/**
* RPC server.
*/
func ServeSurfstoreServer(hostAddr string, surfstoreServer Server) error {
    log.Println("Server started, hostAddr: ", hostAddr)
    rpc.Register(&surfstoreServer)
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", hostAddr)
    if err != nil {
        log.Println("listen error: ", err)
    }
    return http.Serve(l, nil)
}
