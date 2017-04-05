package logging

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	LevelAll = iota
	LevelTrace
	LevelDebug
	LevelInformational
	LevelNotice
	LevelWarning
	LevelError
	LevelCritical
)

const (
	statusReady = iota
	statusRunning
	statusClosing
	statusClosed
)

const (
	defAsyncSize = 1024
)

type Loger interface {
	Name() string
	Open(conf string) error
	Write(msg *Message) (int, error)
	Close() error
	Sync() error
}

type Message struct {
	msgType int64
	message string
}

type Log struct {
	status  int64
	sync    bool
	mutex   sync.Mutex
	logMsg  chan *Message
	sigMsg  chan string
	syncMsg chan struct{}
}

var levelString = make(map[string]int64)
var levelHeadString = make(map[int64]string)
var loggerRegistered = make(map[string]Loger)
var loggerTraced = make(map[string]Loger)

func (l *Log) Critical(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Critical logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelCritical])

	chanMsg := &Message{}
	chanMsg.msgType = LevelCritical

	l.logMessage(chanMsg, logMsg, format, a...)
}

func (l *Log) Error(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Error logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelError])

	chanMsg := &Message{}
	chanMsg.msgType = LevelError

	l.logMessage(chanMsg, logMsg, format, a...)
}

func (l *Log) Warning(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Warning logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelWarning])

	chanMsg := &Message{}
	chanMsg.msgType = LevelWarning

	l.logMessage(chanMsg, logMsg, format, a...)
}
func (l *Log) Notice(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Notice logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelNotice])

	chanMsg := &Message{}
	chanMsg.msgType = LevelNotice

	l.logMessage(chanMsg, logMsg, format, a...)
}

func (l *Log) Info(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Info logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelInformational])

	chanMsg := &Message{}
	chanMsg.msgType = LevelInformational

	l.logMessage(chanMsg, logMsg, format, a...)
}

func (l *Log) Debug(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Debug logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelDebug])

	chanMsg := &Message{}
	chanMsg.msgType = LevelDebug

	l.logMessage(chanMsg, logMsg, format, a...)
}
func (l *Log) Trace(format string, a ...interface{}) {
	if l.status != statusRunning {
		fmt.Printf("Trace logging status not right, status = %d, msg = %s\n",
			l.status, fmt.Sprintf(format, a...))
		return
	}
	now := time.Now()
	logMsg := fmt.Sprintf("[%02d%02d%02d.%06d]%s ",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		levelHeadString[LevelTrace])

	chanMsg := &Message{}
	chanMsg.msgType = LevelTrace

	l.logMessage(chanMsg, logMsg, format, a...)
}
func (l *Log) logMessage(chanMsg *Message, logMsg string, format string, a ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	chanMsg.message = fmt.Sprintf(format, a...)
	chanMsg.message = fmt.Sprintf("%s%s", logMsg, chanMsg.message)

	if l.sync {
		for _, log := range loggerTraced {
			_, err := log.Write(chanMsg)
			if err != nil {
				fmt.Printf("write log message failed. msg = %s\n", chanMsg.message)
			}
		}
		return
	}

	l.logMsg <- chanMsg
}

func (l *Log) Sync() {
	if l.status != statusRunning {
		fmt.Printf("Sync logging status not right, status = %d\n", l.status)
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.sync {
		for _, log := range loggerTraced {
			err := log.Sync()
			if err != nil {
				fmt.Printf("sync message failed\n")
			}
		}
		return
	}
	l.syncMsg <- struct{}{}
}

func (l *Log) StartSync() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if len(loggerTraced) == 0 {
		return fmt.Errorf("log size zero")
	}
	l.status = statusRunning
	l.sync = true

	return nil
}
func (l *Log) StartAsync() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if len(loggerTraced) == 0 {
		return fmt.Errorf("log size zero")
	}
	l.status = statusRunning
	l.sync = false

	l.logMsg = make(chan *Message, defAsyncSize)
	l.sigMsg = make(chan string)
	l.syncMsg = make(chan struct{})

	go l.waitMsg()

	return nil
}

func (l *Log) Start() error {
	return l.StartAsync()
}

func (l *Log) Stop() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.status == statusRunning {
		if l.sync == false {
			l.sigMsg <- "closing"

			// wait for closed
			sig := <-l.sigMsg
			if sig == "closed" {
				l.status = statusClosed

				close(l.logMsg)
				close(l.sigMsg)
			}
		} else {
			// exit logger
			for _, log := range loggerTraced {
				err := log.Close()
				if err != nil {
					fmt.Printf("log close failed.\n")
				}
			}
			l.status = statusClosed
		}
	}
	if l.status == statusClosed {
		for k, _ := range loggerTraced {
			delete(loggerTraced, k)
		}
	}
}

func (l *Log) waitMsg() {
END:
	for {
		select {
		case <-l.syncMsg:
			for _, log := range loggerTraced {
				err := log.Sync()
				if err != nil {
					fmt.Printf("sync message failed\n")
				}
			}
		case msg := <-l.logMsg:
			for _, log := range loggerTraced {
				_, err := log.Write(msg)
				if err != nil {
					fmt.Printf("log write message failed, logger = %s, type = %d, message = %s, err = %s\n",
						log.Name(), msg.msgType, msg.message, err.Error())
				}
			}
			if l.status == statusClosing {
				if len(l.logMsg) == 0 {
					break END
				}
			}
		case sig := <-l.sigMsg:
			if sig == "closing" {
				if len(l.logMsg) == 0 {
					break END
				}
				l.status = statusClosing
			}

		}
	}

	// exit logger
	for _, log := range loggerTraced {
		err := log.Close()
		if err != nil {
			fmt.Printf("log close failed.\n")
		}
	}
	l.sigMsg <- "closed"
}

func NewLogging() (*Log, error) {
	log := &Log{}
	return log, nil
}

func LogLevel(levelStr string) (int64, error) {
	str := strings.ToLower(levelStr)
	if level, ok := levelString[str]; ok {
		return level, nil
	}
	return LevelAll, fmt.Errorf("level %s not found", levelStr)
}

func SetupLog(name string, conf string) (Loger, error) {
	if log, ok := loggerRegistered[name]; ok {
		err := log.Open(conf)
		if err != nil {
			return nil, err
		}
		loggerTraced[name] = log

		return log, nil
	}
	return nil, fmt.Errorf("loger %s not found", name)
}

func Register(log Loger) error {
	name := log.Name()
	if _, ok := loggerRegistered[name]; ok {
		return fmt.Errorf("logger %s exists", name)
	}
	loggerRegistered[name] = log

	return nil
}

func init() {
	levelString["all"] = LevelAll
	levelString["trace"] = LevelTrace
	levelString["debug"] = LevelDebug
	levelString["info"] = LevelInformational
	levelString["notice"] = LevelNotice
	levelString["warn"] = LevelWarning
	levelString["error"] = LevelError
	levelString["critical"] = LevelCritical

	levelHeadString[LevelAll] = "[A]"
	levelHeadString[LevelTrace] = "[T]"
	levelHeadString[LevelDebug] = "[D]"
	levelHeadString[LevelInformational] = "[I]"
	levelHeadString[LevelNotice] = "[N]"
	levelHeadString[LevelWarning] = "[W]"
	levelHeadString[LevelError] = "[E]"
	levelHeadString[LevelCritical] = "[C]"

}
