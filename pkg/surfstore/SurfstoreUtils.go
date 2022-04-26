package surfstore

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Implement the logic for a client syncing with the server here.
func ClientSync(client RPCClient) {
	//1 sync the local to the remote, 2 and from the remote to local, 3 solve the comflict
	localDir := client.BaseDir
	blocksize := client.BlockSize
	//扫描本地localDir
	localMap := make(map[string][]string)
	files, err := ioutil.ReadDir(localDir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.Name() == "index.txt" {
			continue
		}
		filepath := ConcatPath(client.BaseDir, file.Name())
		blockAddr, err := getBlocksAddr(filepath, blocksize)
		if err != nil {
			fmt.Println(err)
		}
		localMap[file.Name()] = blockAddr
	}
	//本地index读一下
	_, err = os.Stat(localDir + "/" + "index.txt")
	if err != nil {
		indexfile, err := os.Create(localDir + "/" + "index.txt")
		if err != nil {
			fmt.Println(err)
		}
		indexfile.Close()
	}
	indexmap, err := LoadMetaFromMetaFile(localDir)
	if err != nil {
		fmt.Println("index.txt read failed", err)
	}
	//覆盖indexmap中的value
	localindexmap, err := overwrite(indexmap, localMap)
	if err != nil {
		fmt.Println("overwrite index.txt failed", err)
	}
	//yuanduan index duyixia
	var remoteindexmap map[string]*FileMetaData
	err = client.GetFileInfoMap(&remoteindexmap)
	if err != nil {
		fmt.Println("get remote fileinfomap failed", err)
	}
	//bendi wenjian upload
	for filename, localfmd := range localindexmap {
		if remotefmd, ok := remoteindexmap[filename]; ok {
			if !compareHash(localfmd.BlockHashList, remotefmd.BlockHashList) { //出现冲突
				// if len(remotefmd.BlockHashList) == 1 && remotefmd.BlockHashList[0] == "0" {
				// 	err = removelocal(client.BaseDir + "/" + filename)
				// 	if err != nil {
				// 		fmt.Println(err)
				// 	}
				// 	localindexmap[filename] = remotefmd
				// } else {
				_, err := updateServer(client, localfmd)
				if err != nil {
					err = downloadServer(client, remotefmd)
					if err != nil {
						fmt.Println(err)
					} else {
						localindexmap[filename] = remotefmd
					}
				}
			}
		} else { //in local but not remote->new file in local
			_, err := updateServer(client, localfmd)
			if err != nil {
				err = downloadServer(client, remotefmd)
				if err != nil {
					fmt.Println(err)
				} else {
					localindexmap[filename] = remotefmd
				}
			}
		}
	}

	// yuanduan chuandao bendi
	for filename, remotefmd := range remoteindexmap {
		if _, ok := localindexmap[filename]; !ok { //delete remote file
			err = downloadServer(client, remotefmd)
			if err != nil {
				fmt.Println(err)
			} else {
				localindexmap[filename] = remotefmd
			}
		}
	}

	// localindexmap := downloadRemote()
	// for filename, fmd := range localindexmap {
	err = WriteMetaFile(localindexmap, client.BaseDir)
	if err != nil {
		fmt.Println(err)
	}
	// }
}

func removelocal(filepath string) error {
	fmt.Println("try to remove file", filepath)
	err := os.Remove(filepath)
	if err != nil {
		return err
	}
	return nil
}

func downloadServer(client RPCClient, remotefmd *FileMetaData) error {
	var ba string
	err := client.GetBlockStoreAddr(&ba)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = removelocal(ConcatPath(client.BaseDir, remotefmd.GetFilename()))
	if err != nil {
		fmt.Println(err)
	}
	if len(remotefmd.BlockHashList) == 0 || len(remotefmd.BlockHashList) > 1 || remotefmd.BlockHashList[0] != "0" {
		f, err := os.Create(ConcatPath(client.BaseDir, remotefmd.GetFilename()))
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()
		for _, h := range remotefmd.BlockHashList {
			var b Block
			err := client.GetBlock(h, ba, &b)
			if err != nil {
				fmt.Println(err)
			}
			_, err = f.Write(b.BlockData)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func updateServer(client RPCClient, localfmd *FileMetaData) (int32, error) {
	var ba string
	err := client.GetBlockStoreAddr(&ba)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	var localdata FileMetaData
	localdata.Filename = localfmd.Filename
	localdata.BlockHashList = localfmd.BlockHashList
	localdata.Version = localfmd.Version
	if len(localfmd.BlockHashList) != 1 || localfmd.BlockHashList[0] != "0" { //upload blocks
		buffer := make([]byte, client.BlockSize)
		f, err := os.Open(ConcatPath(client.BaseDir, localfmd.Filename))
		if err != nil {
			fmt.Println(err)
			return -1, err
		}
		defer f.Close()
		reader := bufio.NewReader(f)
		for {
			n, err := reader.Read(buffer)
			if err != io.EOF && err != nil {
				fmt.Println(err)
				return -1, err
			}
			if n == 0 {
				break
			}
			var b Block
			var s bool
			b.BlockSize = int32(n)
			b.BlockData = buffer[:n]
			err = client.PutBlock(&b, ba, &s)
			if !s || err != nil {
				fmt.Println(err)
				return -1, err
			}
		}
	}

	//update remote indexmap
	var newVersion int32
	err = client.UpdateFile(&localdata, &newVersion)
	if err != nil {
		fmt.Println(err)
		return newVersion, err
	}
	return newVersion, nil
}

func compareHash(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	} else {
		for i := 0; i < len(s1); i++ {
			if s1[i] != s2[i] {
				return false
			}
		}
	}
	return true
}

func overwrite(indexmap map[string]*FileMetaData, localMap map[string][]string) (map[string]*FileMetaData, error) {
	newindexmap := make(map[string]*FileMetaData)
	for filename, hashlist := range localMap {
		var fmd FileMetaData
		fmd.BlockHashList = hashlist
		fmd.Filename = filename
		if val, ok := indexmap[filename]; ok {
			if !compareHash(hashlist, val.BlockHashList) {
				fmd.Version = val.Version + 1
			} else {
				fmd.Version = val.Version
			}
		} else {
			fmd.Version = 1
		}
		newindexmap[filename] = &fmd
	}
	for filename, hashlist := range indexmap {
		if _, ok := localMap[filename]; !ok {
			if len(indexmap[filename].BlockHashList) == 1 && indexmap[filename].BlockHashList[0] == "0" {
				continue
			} else {
				var fmd FileMetaData
				fmd.BlockHashList = []string{"0"}
				fmd.Filename = filename
				fmd.Version = hashlist.Version + 1
				newindexmap[filename] = &fmd
			}
		}
	}
	return newindexmap, nil
}

func getBlocksAddr(filename string, blocksize int) ([]string, error) {
	var res []string
	buffer := make([]byte, blocksize)
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		n, err := reader.Read(buffer)
		if err != io.EOF && err != nil {
			fmt.Println(err)
			return nil, err
		}
		if n == 0 {
			break
		}
		res = append(res, GetBlockHashString(buffer[:n]))
	}
	return res, nil
}
