package xmltest

import (
	"encoding/xml"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

// xml cDATA
type xmlData struct {
	XMLName    xml.Name `xml:"xml"`
	AppID      cDATA    `xml:"appid,omitempty"`       // 1
	CharSet    cDATA    `xml:"charset,omitempty"`     // 2
	DeviceInfo string   `xml:"device_info,omitempty"` // 3
}
type cDATA string

func (c cDATA) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		string `xml:",cDATA"`
	}{string(c)}, start)
}

func TestcDATA(t *testing.T) {
	data := xmlData{
		AppID:      "appid",
		CharSet:    "charset",
		DeviceInfo: "deviceinfo",
	}
	bytes, err := xml.MarshalIndent(data, "", "  ")
	assert.Equal(t, nil, err)
	fmt.Printf("xml:\n%s\n", string(bytes))
}
