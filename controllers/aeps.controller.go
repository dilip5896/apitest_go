package controllers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

/*
=========================
REQUEST STRUCT
=========================
*/
type TwoFARequest struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Device        string  `json:"device"`
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
AES CBC PKCS7 (MATCH CRYPTOJS)
=========================
*/
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func encryptAadhaarCBC(aadhaar, secret string) (string, error) {
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

/*
=========================
AEPS 2FA CONTROLLER
=========================
*/
func TwoFA(c *fiber.Ctx) error {

	// ---------- Parse Request ----------
	var body TwoFARequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// ---------- Aadhaar Encryption ----------
	aadhaar := "634169620558"
	secretKey := "e99816ca42d2c9bce99816ca42d2c9bc"
	if secretKey == "" {
		secretKey = "e99816ca42d2c9bce99816ca42d2c9bc" // fallback
	}

	encryptedAadhaar, err := encryptAadhaarCBC(aadhaar, secretKey)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Encryption failed"})
	}

	// ---------- Parse PID XML ----------
	pidXML := strings.TrimSpace(body.BiometricData.PidData)

	var pid PidData
	if err := xml.Unmarshal([]byte(pidXML), &pid); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Invalid PID XML"})
	}

	// ---------- Handle Device Error ----------
	if pid.Resp.ErrCode == "700" {
		return c.JSON(fiber.Map{
			"success": false,
			"message": pid.Resp.ErrInfo + ", please try again",
		})
	}

	// ---------- Bank ----------
	bankStr := "BANK OF INDIA/508505"
	bankArr := strings.Split(bankStr, "/")
	bankIIN := bankArr[1]

	// ---------- Device Params ----------
	var srno, sysid, ts string
	params := pid.DeviceInfo.Additional.Params

	if len(params) > 0 {
		srno = params[0].Value
	}
	if len(params) > 1 {
		sysid = params[1].Value
	}
	if len(params) > 2 {
		ts = params[2].Value
	}

	// ---------- Biometric Data ----------
	biometricData := map[string]interface{}{
		"encryptedAadhaar": encryptedAadhaar,
		"dc":               pid.DeviceInfo.Dc,
		"ci":               pid.Skey.Ci,
		"hmac":             pid.Hmac,
		"dpId":             pid.DeviceInfo.DpId,
		"mc":               pid.DeviceInfo.Mc,
		"pidDataType":      pid.Data.Type,
		"sessionKey":       pid.Skey.Data,
		"mi":               pid.DeviceInfo.Mi,
		"rdsId":            pid.DeviceInfo.RdsId,
		"errCode":          pid.Resp.ErrCode,
		"errInfo":          pid.Resp.ErrInfo,
		"fCount":           pid.Resp.FCount,
		"fType":            pid.Resp.FType,
		"iCount":           0,
		"iType":            0,
		"pCount":           0,
		"pType":            0,
		"srno":             srno,
		"sysid":            sysid,
		"ts":               ts,
		"pidData":          pid.Data.Data,
		"qScore":           pid.Resp.QScore,
		"nmPoints":         pid.Resp.NmPoints,
		"rdsVer":           pid.DeviceInfo.RdsVer,
	}

	// ---------- Payload ----------
	payload := map[string]interface{}{
		"latitude":      body.Latitude,
		"longitude":     body.Longitude,
		"type":          "DAILY_LOGIN",
		"bankiin":       bankIIN,
		"biometricData": biometricData,
	}
	fmt.Println("payload", payload)

	jsonData, _ := json.Marshal(payload)

	// ---------- InstantPay API ----------
	req, _ := http.NewRequest(
		"POST",
		"https://api.instantpay.in/fi/aeps/outletLogin",
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ipay-Auth-Code", "1")
	req.Header.Set("X-Ipay-Client-Id", "YWY3OTAzYzNlM2ExZTJlOTgVFISIPO97NkBx0nETrpE=")
	req.Header.Set("X-Ipay-Client-Secret", "b6bbfe40ecee5cf96473eaa00f2f2f0742c0f8f63a96e72a13493a8fa3af4a8f")
	req.Header.Set("X-Ipay-Endpoint-Ip", "2405:201:600b:6b:5927:35e4:e336:83ca")
	req.Header.Set("X-Ipay-Outlet-Id", "381229")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "InstantPay API error"})
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
