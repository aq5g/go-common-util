package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"ttutil/filenotify"
)

func main() {
	flag.Parse()
	go watchTargetContents()
	for {
		time.Sleep(time.Hour)
		// do something
	}
}

const (
	targetFile string = "target.json"
)

var (
	targetContents = flag.String("target", "", "Path to if_mib target Contents. like /etc/abc/")
)

type TargetCommunity struct {
	IfMib []IfMib `json:"if_mib"`
}
type IfMib struct {
	Target    string `json:"target"`
	Community string `json:"community"`
}

// watch 目录
func watchTargetContents() {
	if *targetContents == "" {
		return
	}

	tmp, err := filenotify.NewFileNotify(*targetContents)
	if err != nil {
		panic(err)
	}
	defer tmp.Close()
	tmp.StartNotify()
	for {
		// 这里麻烦的原因：
		// 场景：在k8s中，会有一些创建、重命名、删除的操作
		// 导致更新configmap会有多个event
		// 所以是为了一次将所有event都读出来
		<-tmp.ReadChan
		afterr := time.After(2 * time.Second)
	cycle:
		for {
			select {
			case <-afterr:
				break cycle
			case <-tmp.ReadChan:
			}
		}
		err = loadTargetCommunity(*targetContents + targetFile)
		if err != nil {
			log.Println("Error parsing TargetCommunity file: ", err)
		}
	}
}

// 加载target和团体字的映射
// 生成一系列ifmib module的copy
func loadTargetCommunity(filepath string) error {
	// -- 文件是否存在
	if _, err := os.Stat(filepath); err == nil {
		data, err := os.ReadFile(filepath)
		if err != nil {
			return err
		}
		tmpVar := new(TargetCommunity)
		err = json.Unmarshal(data, tmpVar)
		// do something
		if err != nil {
			return err
		}
		return nil
	} else if os.IsNotExist(err) {
		log.Println("msg", filepath+" is not exist")
		return nil
	} else {
		return err
	}
}
