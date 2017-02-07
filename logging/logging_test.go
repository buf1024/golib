package logging

import (
	"testing"
	"time"
)

func TestLogging(t *testing.T) {
	log, err := NewLogging()
	if err != nil {
		t.Errorf("NewLogging failed. err = %s\n", err.Error())
		t.FailNow()
	}
	_, err = SetupLog("file",
		`{"prefix":"hello", "filedir":"./", "level":0, "switchsize":0, "switchtime":0}`)
	if err != nil {
		t.Errorf("setup file logger failed. err = %s\n", err.Error())
		t.FailNow()
	}
	_, err = SetupLog("console", `{"level":0}`)
	if err != nil {
		t.Errorf("setup file logger failed. err = %s\n", err.Error())
		t.FailNow()
	}

	running := 86400 * 2

	log.Start()
	for {
		log.Trace("trace\n")
		log.Debug("debug\n")
		log.Info("info\n")
		log.Notice("notice\n")
		log.Warning("warning\n")
		log.Error("error\n")
		log.Critical("critical\n")
		log.Sync()
		time.Sleep(6 * time.Second)
		running--
		if running <= 0 {
			break
		}
	}
	log.Stop()
}
