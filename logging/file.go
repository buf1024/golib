package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type fileLogger struct {
	Prefix     string `json:"prefix"`
	FileDir    string `json:"filedir"`
	Level      int64  `json:"level"`
	SwitchSize int64  `json:"switchsize"`
	SwitchTime int64  `json:"switchtime"`

	status bool

	file      *os.File
	fileDate  int64
	fileName  string
	fileSize  int64
	fileIndex int64
}

func (f *fileLogger) logSwitch() error {
	n := time.Now()
	if f.file == nil {
		f.fileName = fmt.Sprintf("%s%s_%d_%04d%02d%02d_%d.log.tmp",
			f.FileDir, f.Prefix, os.Getpid(), n.Year(), n.Month(), n.Day(), f.fileIndex)
		f.fileDate = time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.Local).Unix()

		var err error
		f.file, err = os.OpenFile(f.fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		f.fileIndex = 0
		f.fileSize = 0

		return nil
	}
	switchFlag := false
	curDate := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.Local).Unix()

	if f.SwitchTime >= 0 {
		swt := 86400 + f.SwitchTime
		cur := n.Unix()
		if cur-f.fileDate >= swt {
			switchFlag = true
		}
	}
	if f.SwitchSize > 0 {
		if f.fileSize >= f.SwitchSize {
			switchFlag = true
		}
	}

	if switchFlag {
		f.Close()

		if curDate == f.fileDate {
			f.fileIndex++
			return f.logSwitch()
		}
		f.fileIndex = 0
		return f.logSwitch()
	}
	return nil
}

func (f *fileLogger) Name() string {
	return "file"
}

//`{"prefix":"hello", "filedir":"./", "level":0, "switchsize":1024, "switchtime":86400}`)
func (f *fileLogger) Open(conf string) error {
	err := json.Unmarshal([]byte(conf), f)
	if err != nil {
		return err
	}

	if f.Prefix == "" {
		return fmt.Errorf("prefix is empty")
	}
	if f.FileDir == "" {
		return fmt.Errorf("file dir is empty")
	}
	if f.Level < 0 || f.Level > LevelCritical {
		return fmt.Errorf("level must between(%d ~ %d)", LevelAll, LevelCritical)
	}

	f.FileDir = filepath.Dir(f.FileDir)
	if !strings.HasSuffix(f.FileDir, string(filepath.Separator)) {
		f.FileDir += string(filepath.Separator)
	}
	return f.logSwitch()
}
func (f *fileLogger) Write(msg *Message) (int, error) {
	n, err := 0, error(nil)
	if f.file != nil {
		if msg.msgType >= f.Level {
			n, err = f.file.Write([]byte(msg.message))
			if err != nil {
				return n, err
			}
			f.fileSize += int64(n)
			err = f.logSwitch()
			if err != nil {
				return 0, err
			}
			return n, err
		}
	}
	return n, err
}
func (f *fileLogger) Close() error {
	if f.file != nil {
		err := f.file.Close()
		if err != nil {
			return err
		}
		index := strings.Index(f.fileName, ".tmp")
		if index > 0 {
			newPath := f.fileName[:index]
			os.Rename(f.fileName, newPath)
		}
		f.file = nil
	}
	return nil
}
func (f *fileLogger) Sync() error {
	if f.file != nil {
		return f.file.Sync()
	}
	return nil
}

func init() {

	f := &fileLogger{}
	Register(f)
}
