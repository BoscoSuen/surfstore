# Surfstore

This is the starter code for Module 3: Surfstore.  Before you get started, make
sure you understand the following 2 things about Go. (These will also be
covered in class and in discussions)

1. Interfaces: They are named collections of method signatures. Here are some good resources to understand interfaces in Go:
    a. https://gobyexample.com/interfaces
    b. https://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go

2. RPC: You should know how to write RPC servers and clients in Go. The [online documentation](https://golang.org/pkg/net/rpc/) of the *rpc* package is a good resource. 

## Data Types

Recall from the module write-up the following things:

1. The SurfStore service is composed of two services: BlockStore and MetadataStore 
2. A file in SurfStore is broken into an ordered sequence of one or more blocks which are stored in the BlockStore.
3. The MetadataStore maintains the mapping of filenames to hashes of these blocks (and versions) in a map.

The starter code defines the following types for your usage in `SurfstoreInterfaces.go`:

```go
type Block struct {
	BlockData []byte
	BlockSize int
}

type FileMetaData struct {
	Filename      string
	Version       int
	BlockHashList []string
}
```

## Surfstore Interface

`SurfstoreInterfaces.go` also contains interfaces for the BlockStore and the MetadataStore:

```go
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
```

The `Surfstore` interface then glues these two together and is also present in `SurfstoreInterfaces.go`.

```go
type Surfstore interface {
	MetaStoreInterface
	BlockStoreInterface
}
```

## Server

`BlockStore.go` provides a skeleton implementation of the `BlockStoreInterface`
and `MetaStore.go` provides a skeleton implementation of the
`MetaStoreInterface` **You must implement the methods in these 2 files which
have `panic("todo")` as their body.**

`SurfstoreServer.go` should then put everything together to provide a complete
implementation of the `Surfstore` interface. **You must implement the methods
in this file which have `panic("todo")` as their body.** (Hint: You have
already implemented these for the `BlockStore` and the `Metastore`, you just
need to call them appropriately. )

`SurfstoreServer.go` also has a method `ServeSurfstoreServer` **which you must
implement**. It should register the `Server` instance passed to it and start
listening for connections from clients. 

## Client

`SurfstoreRPCClient.go` provides the rpc client stub for the surfstore rpc
server. **You must implement the methods in this file which have
`panic("todo")` as their body.** (Hint: one of them has been implemented for
you) 

`SurfstoreClientUtils.go` also has the following method which **you need to
implement** for the sync logic of clients:

```go

/*
Implement the logic for a client syncing with the server here.
*/
func ClientSync(client RPCClient) {
	panic("todo")
}
```

## Setup

You will need to setup your runtime environment variables so that you can build
your code and also use the executables that will be generated.

1. If you are using a Mac, open `~/.bash_profile` or if you are using a
unix/linux machine, open `~/.bashrc`. Then add the following:

```
export GOPATH=<path to starter code>
export PATH=$PATH:$GOPATH/bin
```

2. Run `source ~/.bash_profile` or `source ~/.bashrc`

## Usage

1. Only after you have implemented all the methods and completed the `Setup`
steps, run the `build.sh` script provided with the starter code. This should
create 2 executables in the `bin` folder inside your starter code directory.

```shell
> ./build.sh
> ls bin
SurfstoreClientExec SurfstoreServerExec
```

2. Run your server using the script provided in the starter code.

```shell
./run-server.sh
```

3. From a new terminal (or a new node), run the client using the script
provided in the starter code (if using a new node, build using step 1 first).
Use a base directory with some files in it.

```shell
> mkdir dataA
> cp ~/pic.jpg dataA/ 
> ./run-client.sh server_addr:port dataA 4096
```

This would sync pic.jpg to the server hosted on `server_addr:port`, using
`dataA` as the base directory, with a block size of 4096 bytes.

4. From another terminal (or a new node), run the client to sync with the
server. (if using a new node, build using step 1 first)

```shell
> ls dataB/
> ./run-client.sh server_addr:port dataB 4096
> ls dataB/
pic.jpg index.txt
```

We observe that pic.jpg has been synced to this client.
