package surfstore

import (
	context "context"
	"errors"
	"sync"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type MetaStore struct {
	suo            sync.Mutex
	FileMetaMap    map[string]*FileMetaData
	BlockStoreAddr string
	UnimplementedMetaStoreServer
}

func (m *MetaStore) GetFileInfoMap(ctx context.Context, _ *emptypb.Empty) (*FileInfoMap, error) {
	m.suo.Lock()
	defer m.suo.Unlock()
	res := &FileInfoMap{FileInfoMap: m.FileMetaMap}
	return res, nil
}

func (m *MetaStore) UpdateFile(ctx context.Context, fileMetaData *FileMetaData) (*Version, error) {
	m.suo.Lock()
	defer m.suo.Unlock()
	filename := fileMetaData.Filename
	// curvision := fileMetaData.Version
	oldData, flag := m.FileMetaMap[filename]
	if !flag { //没找到老文件
		//add new filename and filetext to FileMetaMap
		fileMetaData.Version = 1
		m.FileMetaMap[filename] = fileMetaData
		return &Version{Version: 1}, nil
	} else { //找到了！哈哈哈
		//要先判断版本
		if fileMetaData.Version != (oldData.Version + 1) { //版本号不合法
			return nil, errors.New("your version is wrong")
		} else { //delete and update is the same action
			// m.FileMetaMap[filename] = fileMetaData
			// return &Version{Version: curvision}, nil
			if len(fileMetaData.BlockHashList) == 1 && fileMetaData.BlockHashList[0] == "0" { // delete remote file
				if len(oldData.BlockHashList) == 0 || len(oldData.BlockHashList) > 1 || oldData.BlockHashList[0] != "0" { // not deleted yet
					m.FileMetaMap[fileMetaData.Filename].Version += 1
					m.FileMetaMap[fileMetaData.Filename].BlockHashList = []string{"0"}
					return &Version{Version: m.FileMetaMap[fileMetaData.Filename].Version}, nil
				} else { // already dead
					return &Version{Version: oldData.Version + 1}, errors.New("File already deleted before!")
				}
			} else {
				if oldData.Version+1 == fileMetaData.Version {
					m.FileMetaMap[fileMetaData.Filename].Version = fileMetaData.Version
					m.FileMetaMap[fileMetaData.Filename].BlockHashList = fileMetaData.GetBlockHashList()
					return &Version{Version: fileMetaData.Version}, nil
				} else {
					return &Version{Version: oldData.Version}, errors.New("your version is wrong")
				}
			}
		}
	}

}

func (m *MetaStore) GetBlockStoreAddr(ctx context.Context, _ *emptypb.Empty) (*BlockStoreAddr, error) {
	m.suo.Lock()
	defer m.suo.Unlock()
	res := &BlockStoreAddr{Addr: m.BlockStoreAddr}
	return res, nil
}

// This line guarantees all method for MetaStore are implemented
var _ MetaStoreInterface = new(MetaStore)

func NewMetaStore(blockStoreAddr string) *MetaStore {
	return &MetaStore{
		FileMetaMap:    map[string]*FileMetaData{},
		BlockStoreAddr: blockStoreAddr,
	}
}
