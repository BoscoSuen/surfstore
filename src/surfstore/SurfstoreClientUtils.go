package surfstore

import (
    "crypto/sha256"
    "encoding/hex"
    "io/ioutil"
    "fmt"
    "log"
    "io"
    "math"
    "os"
    "strings"
    "strconv"
)

/*
 * Implement the logic for a client syncing with the server here.
 * first scan the base directory, and for each file, compute that file’s hash list.
 * The client should then consult the local index file and compare the results, 
 * to see whether (1) there are now new files in the base directory that aren’t 
 * in the index file, or (2) files that are in the index file, but have changed 
 * since the last time the client was executed (i.e., the hash list is different).
 */
func ClientSync(client RPCClient) {
    baseDir, readErr := ioutil.ReadDir(client.BaseDir)
    if readErr != nil {
        log.Println("Read client base directory error: ", readErr)
    }
    dirMap := make(map[string]os.FileInfo)
    for _, file := range baseDir {
        dirMap[file.Name()] = file
    }
    
    // Check basic index.txt file
    // index.txt format example: File1.dat,3,h0 h1 h2 h3
    indexFilePath := client.BaseDir + "/index.txt"
    if _, indexFileErr := os.Stat(indexFilePath); os.IsNotExist(indexFileErr) {
        file, _ := os.Create(indexFilePath)
        defer file.Close()
    }

    indexMap := make(map[string]int)

    indexFileInfoMap := make(map[string]FileMetaData)

    indexFile, _ := ioutil.ReadFile(indexFilePath)
    indexLines := strings.Split(string(indexFile), "\n")
    for i, line := range indexLines {
        if line == "" {
            continue
        }
        fileMetaData := encode(string(line))
        indexFileInfoMap[fileMetaData.Filename] = fileMetaData
        indexMap[fileMetaData.Filename] = i
    }

    // Iterate baseDir files and sync with index.txt, update file status in a new map
    clientFileInfoMap := localSync(client, indexFileInfoMap, &indexMap , dirMap , &indexLines)

    var succ bool
    serverFileInfoMap := make(map[string]FileMetaData)
    getInfoMapErr := client.GetFileInfoMap(&succ, &serverFileInfoMap)
    if getInfoMapErr != nil {
        log.Println("Get file info map from server error: ", getInfoMapErr)
    }

    // Upload updated file to server
    for fileName, info := range clientFileInfoMap {
        // Check if the server has the file
        if _, ok := serverFileInfoMap[fileName]; ok {
            serverFileMetaData := serverFileInfoMap[fileName]
            clientFileMetaData := info.FileMetaData
            if clientFileMetaData.Version == serverFileMetaData.Version && info.Status == Unchanged {
                continue
            } else if (clientFileMetaData.Version > serverFileMetaData.Version) || 
                      (clientFileMetaData.Version == serverFileMetaData.Version && info.Status == Modified) {
                // Server side file is old, or version is same and update only if file is modified
                updateServerFile(client, clientFileMetaData, indexMap, &indexLines, info.Status)
            } else {
                // Client side file is old, or the file version is the same, update the client file.
                updateClientFile(client, serverFileMetaData, indexMap, &indexLines)
            }
        } else {
            // upload file to the server
            upload(client, info.FileMetaData, indexMap, &indexLines)
        }
    }
    
    // Only download NEW files from server
    for fileName, serverFileMetaData := range serverFileInfoMap {
        if _, ok := clientFileInfoMap[fileName]; !ok {
            if _, okay := indexMap[fileName]; okay {
                // The file is deleted locally, check the version
                deletedFileMetaData := encode(indexLines[indexMap[fileName]])
                if deletedFileMetaData.Version > serverFileMetaData.Version {
                    // Upload deleted fileMetaData to the server
                    updateServerFile(client, deletedFileMetaData, indexMap, &indexLines, Deleted)
                } else {
                    // Download file from the server
                    updateClientFile(client, serverFileMetaData, indexMap, &indexLines)
                }
            } else {
                // Download brand new file.
                line, err := download(client, fileName, serverFileMetaData)
                if err != nil {
                    log.Println("Download file from server failed: ", err)
                }
                indexLines = append((indexLines), line)
            }
        }
    }

    // Update index.txt file
    updatedIndexFile := ""
    for _, indexLine := range indexLines {
        if indexLine == "" {
            continue
        }
        updatedIndexFile += indexLine + "\n"
    }
    
    err := ioutil.WriteFile(indexFilePath, []byte(updatedIndexFile), 0755)
    if err != nil {
        log.Println("Updating index.txt file failed: ", err)
    }
}


/**
* Encode line in the index.txt file.
*/
func encode(line string) FileMetaData {
    var fileMetaData FileMetaData
    tokens := strings.Split(line, ",")
    if len(tokens) != 3 {
        log.Println("Token size is not 3")
    }
    fileMetaData.Filename = tokens[0]
    fileMetaData.Version, _ = strconv.Atoi(tokens[1])
    hashListStr := tokens[2]
    hashListTokens := strings.Fields(hashListStr)
    var hashList []string
    for _, hashListToken := range hashListTokens {
        hashList = append(hashList, hashListToken)
    }
    fileMetaData.BlockHashList = hashList
    return fileMetaData
}

/**
* Sync local dir with index.txt
* If file is deleted, set the hashlist to "0"
* If file has been modified, update fileInfo, after comparing with server files, then updating.
*/
func localSync(client RPCClient, indexFileInfoMap map[string]FileMetaData, indexMap *map[string]int, dirMap map[string]os.FileInfo, indexLines *[]string) (map[string]FileInfo) {
    // Check deleted files
    checkDeletedFiles(indexFileInfoMap, *indexMap, dirMap, indexLines)
    
    localMap := make(map[string]FileInfo)
    // Record file status
    for fileName, f := range dirMap {
        if fileName == "index.txt" {
            continue
        }

        file, openErr := os.Open(client.BaseDir + "/" + fileName)
        if openErr != nil {
            log.Println("Open file Error: ", openErr)
        }
        fileSize := f.Size()
        numBlock := int(math.Ceil(float64(fileSize) / float64(client.BlockSize)))

        // Check if file is a new file that is not recorded in index.txt or modified or unchanged
        var info FileInfo

        if fileMetaData, ok := indexFileInfoMap[fileName]; ok {
            // index.txt has the file record
            changed, hashList := getHashList(file, fileMetaData, numBlock, client.BlockSize)
            info.FileMetaData.Filename = fileName
            info.FileMetaData.Version = fileMetaData.Version
            hashStr := ""
            for i, hash := range hashList {
                info.FileMetaData.BlockHashList = append(info.FileMetaData.BlockHashList, hash)
                hashStr += hash
                if i != len(hashList) - 1 {
                    hashStr += " " 
                }
            }
            if changed {
                info.Status = Modified
                // update index.txt
                index := (*indexMap)[fileName]
                (*indexLines)[index] = fileName + "," + strconv.Itoa(fileMetaData.Version) + "," + hashStr
            } else {
                info.Status = Unchanged
            }
        } else {
            // index.txt does not have the file record, i.e, no such a FileMetaData recorded.
            var metaData FileMetaData
            _, hashList := getHashList(file, metaData, numBlock, client.BlockSize)
            info.FileMetaData.Filename = fileName
            info.FileMetaData.Version = 1
            hashStr := ""
            for idx, hash := range hashList {
                info.FileMetaData.BlockHashList = append(info.FileMetaData.BlockHashList, hash)
                hashStr += hash
                if idx != len(hashList) - 1 {
                    hashStr += " "
                }
            }
            info.FileMetaData.BlockHashList = hashList
            info.Status = New

            *indexLines = append((*indexLines), fileName + "," + strconv.Itoa(info.FileMetaData.Version) + "," + hashStr)

            // Add new indexing in the indexMap
            (*indexMap)[fileName] = len(*indexLines) - 1
        }

        localMap[fileName] = info
    }
    return localMap
}

/**
* If file has been deleted, change its hashList to "0"
*/
func checkDeletedFiles(indexFileInfoMap map[string]FileMetaData, indexMap map[string]int, dirMap map[string]os.FileInfo, indexLines *[]string) {
    for fileName, metadata := range indexFileInfoMap {
        if _, ok := dirMap[fileName]; !ok {
            // file recorded in index.txt, but deleted in dir, version will increase and hashlist update to "0"
            index := indexMap[fileName]
            if len(metadata.BlockHashList) == 1 && metadata.BlockHashList[0] == "0" {
                (*indexLines)[index] = metadata.Filename + "," + strconv.Itoa(metadata.Version) + ",0"
            } else {
                (*indexLines)[index] = metadata.Filename + "," + strconv.Itoa(metadata.Version + 1) + ",0"
            }
        }
    }
}

/**
* Generate hashList from file data blocks.
*/
func getHashList(file *os.File, fileMetaData FileMetaData, numBlock int, blockSize int) (bool, []string) {
    hashList := make([]string, numBlock)
    var changed bool
    for i := 0; i < numBlock; i++ {
        // For each block, generate the hashList
        buf := make([]byte, blockSize)
        n, e := file.Read(buf)
        if e != nil {
            log.Println("read error when getting hashList: ", e)
        }
        // Trim the buf
        buf = buf[:n]

        hash := sha256.New()
        hash.Write(buf)
        hashBytes := hash.Sum(nil)
        hashCode := hex.EncodeToString(hashBytes)
        hashList[i] = hashCode
        if i >= len(fileMetaData.BlockHashList) || hashCode != fileMetaData.BlockHashList[i] {
            changed = true
        }
    }
    if numBlock != len(fileMetaData.BlockHashList) {
        changed = true
    }
    return changed, hashList
}

/**
* Upload file to the server
* UpdateFile and PutBlock
*/
func upload(client RPCClient, fileMetaData FileMetaData, indexMap map[string]int, indexLines *[]string) error {
    // Update file blocks
    var err error

    filePath := client.BaseDir + "/" + fileMetaData.Filename
    if _, e := os.Stat(filePath); os.IsNotExist(e) {
        // local file has been deleted, do not need to push blocks
        err = client.UpdateFile(&fileMetaData, &fileMetaData.Version)
        if err != nil {
            log.Println("Update file failed: ", err)
        }
        return err
    }

    file, openErr := os.Open(filePath)
    if openErr != nil {
        log.Println("Open file Error: ", openErr)
    }

    defer file.Close()

    f, _ := os.Stat(filePath)
    numBlock := int(math.Ceil(float64(f.Size()) / float64(client.BlockSize)))

    // Put Block
    for i := 0; i < numBlock; i++ {
        var block Block
        block.BlockData = make([]byte, client.BlockSize)
        n, readErr := file.Read(block.BlockData)
        if readErr != nil && readErr != io.EOF{
            log.Println("Read file error: ", readErr)
        }
        block.BlockSize = n 
        // Trim the blockData
        block.BlockData = block.BlockData[:n]

        var succ bool
        err = client.PutBlock(block, &succ)
        if err != nil {
            log.Println("Put block failed: ", err)
        }
    } 

    // Update file
    err = client.UpdateFile(&fileMetaData, &fileMetaData.Version)
    if err != nil {
        log.Println("Update file failed: ", err)
        // Update file failed, download from the server
        serverFileInfoMap := make(map[string]FileMetaData)
        var succ bool
        client.GetFileInfoMap(&succ, &serverFileInfoMap)
        updateClientFile(client, serverFileInfoMap[fileMetaData.Filename], indexMap, indexLines)
    }
    return err
}

/**
* Upload new client side file to the server.
*/
func updateServerFile(client RPCClient, clientFileMetaData FileMetaData, indexMap map[string]int, indexLines *[]string, status State) {
    // Update file
    // If client file has updated, version should plus 1.
    if status == Modified {
        clientFileMetaData.Version += 1
        // index.txt should update the version
        index := indexMap[clientFileMetaData.Filename]
        line := (*indexLines)[index]
        (*indexLines)[index] = line[:strings.Index(line, ",")] + "," + strconv.Itoa(clientFileMetaData.Version) + "," + line[strings.LastIndex(line, ",") + 1:]
    }

    err := upload(client, clientFileMetaData, indexMap, indexLines)

    if err != nil {
        log.Println("Upload file failed: ", err)
    }
}

/**
* Download file from server.
* If the file exists in the client dir, overwrite the file, otherwise create the file.
*/
func download(client RPCClient, fileName string, fileMetaData FileMetaData) (string, error) {
    filePath := client.BaseDir + "/" + fileName
    if _, e := os.Stat(filePath); os.IsNotExist(e) {
        os.Create(filePath)
    } else {
        os.Truncate(filePath, 0)    // Clean the current file.
    }

    if len(fileMetaData.BlockHashList) == 1 && fileMetaData.BlockHashList[0] == "0" {
        // file in the server has been deleted
        err := os.Remove(filePath)
        if err != nil {
            log.Println("Cannot remove file: ", err)
        }
        line := fileMetaData.Filename + "," + strconv.Itoa(fileMetaData.Version) + ",0"
        return line, err
    }

    file, _ := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)    // Add file access mode.
    defer file.Close()

    hashStr := ""
    var err error
    for i, hash := range fileMetaData.BlockHashList {
        var blockData Block
        err = client.GetBlock(hash, &blockData)
        if err != nil {
            log.Println("Get block failed: ", err)
        }

        data := string(blockData.BlockData)

        _, err = io.WriteString(file, data)
        if err != nil {
            log.Println("Write file failed: ", err)
        }

        hashStr += hash
        if i != len(fileMetaData.BlockHashList) - 1 {
            hashStr += " "
        }
    }
    line := fileMetaData.Filename + "," + strconv.Itoa(fileMetaData.Version) + "," + hashStr
    return line, err
}

/**
* Fetch file from server and update the client side file.
* Index.txt also needs updating.
*/
func updateClientFile(client RPCClient, serverFileMetaData FileMetaData, indexMap map[string]int, indexLines *[]string) {
    line, err := download(client, serverFileMetaData.Filename, serverFileMetaData)

    if err != nil {
        log.Println("Download file from server failed: ", err)
    }

    index := indexMap[serverFileMetaData.Filename]
    (*indexLines)[index] = line
}

/*
* Helper function to print the contents of the metadata map.
*/
func PrintMetaMap(metaMap map[string]FileMetaData) {

    fmt.Println("--------BEGIN PRINT MAP--------")

    for _, filemeta := range metaMap {
        fmt.Println("\t", filemeta.Filename, filemeta.Version, filemeta.BlockHashList)
    }

    fmt.Println("---------END PRINT MAP--------")

}
