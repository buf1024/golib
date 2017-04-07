package xmltest

import (
	"encoding/xml"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

// xml xmlMap
type xmlMap map[string]interface{}

func (m xmlMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	//start.Name = xml.Name{
	//	Space: "",
	//	Local: "map",
	//}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range m {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}
	return nil
}
func TestMap(t *testing.T) {
	data := xmlMap{
		"AppID":      "appid",
		"CharSet":    "charset",
		"DeviceInfo": "deviceinfo",
	}
	bytes, err := xml.MarshalIndent(data, "", "  ")
	assert.Equal(t, nil, err)
	fmt.Printf("xml:\n%s\n", string(bytes))
}
