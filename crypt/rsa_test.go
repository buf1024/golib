package crypt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"

	"encoding/base64"
	"fmt"
)

func TestRsaCrypt(t *testing.T) {
	bs, err := ioutil.ReadFile("1.key")
	if err != nil {
		t.Fatalf("read private key file failed, err = %s", err.Error())
	}
	block, _ := pem.Decode(bs)
	if block == nil {
		t.Fatalf("decode private key failed")
	}
	privt, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse private key failed, err=%s", err)
	}
	testData := "hellow, afsaf, what the fuck ??????,世界还是美好的"

	testEncData, err := PrivateEncrypt(privt, []byte(testData))
	if err != nil {
		t.Fatalf("private key encrypt failed, err=%s", err)
	}
	fmt.Printf("orig data=%s\nenc data=%s\n",
		testData, base64.StdEncoding.EncodeToString(testEncData))

	bs, err = ioutil.ReadFile("1.cer")
	if err != nil {
		t.Fatalf("read certifacte file failed, err = %s", err.Error())
	}
	block, _ = pem.Decode(bs)
	if block == nil {
		t.Fatalf("decode certifacte key failed")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse certifacte key failed, err=%s", err)
	}
	pub := cert.PublicKey.(*rsa.PublicKey)
	testDecData, err := PublicDecrypt(pub, testEncData)
	if err != nil {
		t.Fatalf("public decrypt failed")
	}
	fmt.Printf("dec data = %s\n", string(testDecData))
}
