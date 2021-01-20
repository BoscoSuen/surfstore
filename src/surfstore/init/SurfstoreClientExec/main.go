package main

import (
    "fmt"
    "os"
    "strconv"
    "surfstore"
)

func main() {
    if len(os.Args) < 4 {
        fmt.Println("Usage: ./run-client host:port baseDir blockSize")
        os.Exit(1)
    }

    hostPort := os.Args[1]
    baseDir := os.Args[2]
    blockSize, err := strconv.Atoi(os.Args[3])
    if err != nil {
        fmt.Println("Usage: ./run-client host:port baseDir blockSize")
    }
    rpcClient := surfstore.NewSurfstoreRPCClient(hostPort, baseDir, blockSize)
    surfstore.ClientSync(rpcClient)
}
