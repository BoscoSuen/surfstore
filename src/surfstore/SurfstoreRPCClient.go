package surfstore

import (
    "net/rpc"
)

type RPCClient struct {
    ServerAddr string
    BaseDir    string
    BlockSize  int
}

func (surfClient *RPCClient) GetBlock(blockHash string, block *Block) error {
    // connect to the server
    conn, e := rpc.DialHTTP("tcp", surfClient.ServerAddr)
    if e != nil {
        return e
    }

    // perform the call
    e = conn.Call("Server.GetBlock", blockHash, block)
    if e != nil {
        conn.Close()
        return e
    }

    // close the connection
    return conn.Close()
}

func (surfClient *RPCClient) PutBlock(block Block, succ *bool) error {
    // connect to the server
    conn, e := rpc.DialHTTP("tcp", surfClient.ServerAddr)
    if e != nil {
        return e
    }

    // perform the call
    e = conn.Call("Server.PutBlock", block, succ)
    if e != nil {
        conn.Close()
        return e
    }

    // close the connection
    return conn.Close()
}

func (surfClient *RPCClient) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
    // connect to the server
    conn, e := rpc.DialHTTP("tcp", surfClient.ServerAddr)
    if e != nil {
        return e
    }

    // perform the call
    e = conn.Call("Server.HasBlocks", blockHashesIn, blockHashesOut)
    if e != nil {
        conn.Close()
        return e
    }

    // close the connection
    return conn.Close()
}

func (surfClient *RPCClient) GetFileInfoMap(succ *bool, serverFileInfoMap *map[string]FileMetaData) error {
    // connect to the server
    conn, e := rpc.DialHTTP("tcp", surfClient.ServerAddr)
    if e != nil {
        return e
    }

    // perform the call
    e = conn.Call("Server.GetFileInfoMap", succ, serverFileInfoMap)
    if e != nil {
        conn.Close()
        return e
    }

    // close the connection
    return conn.Close()
}

func (surfClient *RPCClient) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
    // connect to the server
    conn, e := rpc.DialHTTP("tcp", surfClient.ServerAddr)
    if e != nil {
        return e
    }

    // perform the call
    e = conn.Call("Server.UpdateFile", fileMetaData, latestVersion)
    if e != nil {
        conn.Close()
        return e
    }

    // close the connection
    return conn.Close()
}

var _ Surfstore = new(RPCClient)

// Create an Surfstore RPC client
func NewSurfstoreRPCClient(hostPort, baseDir string, blockSize int) RPCClient {

    return RPCClient{
        ServerAddr: hostPort,
        BaseDir:    baseDir,
        BlockSize:  blockSize,
    }
}
