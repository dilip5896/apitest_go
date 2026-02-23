package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"html"
	"io"
	"regexp"
	"strings"
)

/*
	=========================

REQUEST STRUCT
=========================
*/
func CleanPidXML(pid string) string {

	// Trim spaces
	pid = strings.TrimSpace(pid)

	// HTML unescape (&lt; &gt;)
	pid = html.UnescapeString(pid)

	// Remove CDATA
	re := regexp.MustCompile(`<!\[CDATA\[|\]\]>`)
	pid = re.ReplaceAllString(pid, "")

	// Remove BOM if exists
	pid = strings.TrimPrefix(pid, "\ufeff")

	return strings.TrimSpace(pid)
}

type TwoFARequest struct {
	Mobile    string  `json:"mobile"`
	Aadhaar   string  `json:"aadhaar"`
	BankIIN   string  `json:"bankiin"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Device    string  `json:"device"`
	Method    string  `json:"method"`

	// direct pidData
	PidData string `json:"pidData"`

	// nested pidData
	BiometricData struct {
		PidData string `json:"pidData"`
	} `json:"biometricData"`
}
type BalanceRequest struct {
	Mobile    string  `json:"mobile"`
	Aadhaar   string  `json:"aadhaar"`
	BankIIN   string  `json:"bankiin"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Device    string  `json:"device"`
	Method    string  `json:"method"`

	// 👇 supports root pidData
	PidData string `json:"pidData"`

	BiometricData struct {
		PidData string `json:"pidData"`
	} `json:"biometricData"`
}

/*
	=========================

PID XML STRUCT
=========================
*/
type PidData struct {
	XMLName    xml.Name   `xml:"PidData"`
	Resp       Resp       `xml:"Resp"`
	DeviceInfo DeviceInfo `xml:"DeviceInfo"`
	Skey       Skey       `xml:"Skey"`
	Hmac       string     `xml:"Hmac"`
	Data       Data       `xml:"Data"`
}

type Resp struct {
	ErrCode  string `xml:"errCode,attr"`
	ErrInfo  string `xml:"errInfo,attr"`
	FCount   string `xml:"fCount,attr"`
	FType    string `xml:"fType,attr"`
	QScore   string `xml:"qScore,attr"`
	NmPoints string `xml:"nmPoints,attr"`
}

type DeviceInfo struct {
	DpId       string         `xml:"dpId,attr"`
	RdsId      string         `xml:"rdsId,attr"`
	RdsVer     string         `xml:"rdsVer,attr"`
	Mi         string         `xml:"mi,attr"`
	Dc         string         `xml:"dc,attr"`
	Mc         string         `xml:"mc,attr"`
	Additional AdditionalInfo `xml:"additional_info"`
}

type AdditionalInfo struct {
	Params []Param `xml:"Param"`
}

type Param struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Skey struct {
	Ci   string `xml:"ci,attr"`
	Data string `xml:",chardata"`
}

type Data struct {
	Type string `xml:"type,attr"`
	Data string `xml:",chardata"`
}

/*
	=========================

AES CBC PKCS7
=========================
*/
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func EncryptAadhaarCBC(aadhaar, secret string) (string, error) {
	key := []byte(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	plain := pkcs7Pad([]byte(aadhaar), block.BlockSize())
	ciphertext := make([]byte, len(plain))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plain)

	final := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(final), nil
}
