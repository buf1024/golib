package logging

import (
	"encoding/json"
	"github.com/fatih/color"
)

var colorLevel = make(map[int64]color.Attribute)

type consoleLogger struct {
	Level int64 `json:"level"`
}

func (c *consoleLogger) Name() string {
	return "console"
}
func (c *consoleLogger) Open(conf string) error {
	err := json.Unmarshal([]byte(conf), &c)
	return err
}
func (c *consoleLogger) Write(msg *Message) (int, error) {
	n, err := 0, error(nil)
	if msg.msgType >= c.Level {
		n, err = color.New(colorLevel[msg.msgType]).Print(msg.message)
	}
	return n, err
}
func (c *consoleLogger) Close() error {
	return nil
}
func (c *consoleLogger) Sync() error {
	return nil
}

func init() {

	colorLevel[LevelCritical] = color.FgRed
	colorLevel[LevelError] = color.FgHiRed
	colorLevel[LevelWarning] = color.FgMagenta
	colorLevel[LevelNotice] = color.FgBlue
	colorLevel[LevelInformational] = color.FgHiBlue
	colorLevel[LevelDebug] = color.FgGreen
	colorLevel[LevelTrace] = color.FgHiGreen
	colorLevel[LevelAll] = color.FgWhite

	c := &consoleLogger{}
	Register(c)
}
