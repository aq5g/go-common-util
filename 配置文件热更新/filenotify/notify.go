package filenotify

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileNotify struct {
	// 本身支持监听多个路径
	watcher  *fsnotify.Watcher
	isClose  bool
	FilePath string
	ReadChan chan int
}

// 1. 若监听文件，一开始必须存在
// 2. 文件一旦删除，readfile肯定出错
//    并且一旦删除，监听本身不再有效了
func NewFileNotify(filePath string) (*FileNotify, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(filePath)
	if err != nil {
		return nil, err
	}
	return &FileNotify{watcher: watcher, FilePath: filePath, ReadChan: make(chan int, 10)}, nil
}

func (tmp *FileNotify) StartNotify() {
	go func() {
		for {
			select {
			case event, ok := <-tmp.watcher.Events:
				if !ok {
					return
				}
				log.Println("Event:", event)
				tmp.ReadChan <- 1

			case err, ok := <-tmp.watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()
}

func (tmp *FileNotify) ReadFile() ([]byte, error) {
	return os.ReadFile(tmp.FilePath)
}

func (tmp *FileNotify) Close() error {
	if tmp.isClose {
		return nil
	}
	tmp.isClose = true
	close(tmp.ReadChan)
	return tmp.watcher.Close()
}
