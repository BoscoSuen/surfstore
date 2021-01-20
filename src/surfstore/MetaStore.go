package surfstore

import (
    "errors"
)

type MetaStore struct {
    FileMetaMap map[string]FileMetaData
}

/**
* Returns a mapping of the files stored in the SurfStore cloud service, 
* including the version, filename, and hashlist.
*/
func (m *MetaStore) GetFileInfoMap(_ignore *bool, serverFileInfoMap *map[string]FileMetaData) error {
    // Only retrieve the map
    *serverFileInfoMap = m.FileMetaMap
    return nil
}

/**
* Updates the FileInfo values associated with a file stored in the cloud. This method replaces the hash list 
* for the file with the provided hash list ___only if the new version number is exactly one greater___ 
* than the current version number. Otherwise, an error is sent to the client telling them that the version 
* they are trying to store is not right (likely too old) as well as the current value of the file’s version on the server.
*/
func (m *MetaStore) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) (err error) {
    filename := fileMetaData.Filename
    // File may not exist in the metaStore map
    if _, ok := m.FileMetaMap[filename]; ok {
        if fileMetaData.Version - m.FileMetaMap[filename].Version == 1 {
            m.FileMetaMap[filename] = *fileMetaData
            *latestVersion = fileMetaData.Version       // Update the lastest version as the new version.
            return nil
        } else {
            return errors.New("New version number is NOT one greater than current version number")
        }
    } else {
        m.FileMetaMap[filename] = *fileMetaData
        *latestVersion = fileMetaData.Version
        return nil
    }
}

var _ MetaStoreInterface = new(MetaStore)
