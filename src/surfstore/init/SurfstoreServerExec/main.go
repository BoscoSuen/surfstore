package main

import (
    "log"
    "surfstore"
)

func main() {
    serverInstance := surfstore.NewSurfstoreServer()
    log.Fatal(surfstore.ServeSurfstoreServer("localhost:8080", serverInstance))
}
