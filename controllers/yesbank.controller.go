package controllers

import (
	"aepsapi/utils"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

/*
	=========================
	  EXTERNAL REF GENERATOR

=======================
=========================
*/
func Firstapi(c *fiber.Ctx) error {

	// ---------- Parse Body ----------
	var body utils.BalanceRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	fmt.Println("RAW BODY =>", string(c.Body()))

	// ---------- Resolve PID DATA (ROOT OR biometricData) ----------
	pidXML := body.PidData
	if pidXML == "" {
		pidXML = body.BiometricData.PidData
	}

	if pidXML == "" {
		return c.JSON(fiber.Map{
			"success": false,
			"message": "PID data missing",
		})
	}

	// ---------- Basic Validation ----------
	if body.Aadhaar == "" || body.BankIIN == "" || body.Mobile == "" {
		return c.JSON(fiber.Map{
			"success": false,
			"message": "Required fields missing",
		})
	}

	// ---------- Encrypt Aadhaar ----------
	encryptedAadhaar, err := utils.EncryptAadhaarCBC(
		body.Aadhaar,
		"e99816ca42d2c9bce99816ca42d2c9bc",
	)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"message": "Aadhaar encryption failed",
		})
	}

	// ---------- Clean + Parse PID XML ----------
	cleanPid := utils.CleanPidXML(pidXML)

	var pid utils.PidData
	if err := xml.Unmarshal([]byte(cleanPid), &pid); err != nil {
		fmt.Println("PID XML ERROR:", err)
		return c.JSON(fiber.Map{
			"success": false,
			"message": "Invalid PID XML",
		})
	}

	// ---------- Device Error ----------
	if pid.Resp.ErrCode != "0" {
		return c.JSON(fiber.Map{
			"success": false,
			"message": pid.Resp.ErrInfo,
		})
	}

	// ---------- Device Params ----------
	var srno, sysid, ts string
	for _, p := range pid.DeviceInfo.Additional.Params {
		switch p.Name {
		case "srno":
			srno = p.Value
		case "sysid":
			sysid = p.Value
		case "ts":
			ts = p.Value
		}
	}

	// ---------- Biometric Payload ----------
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
		"rdsVer":           pid.DeviceInfo.RdsVer,
		"errCode":          pid.Resp.ErrCode,
		"errInfo":          pid.Resp.ErrInfo,
		"fCount":           pid.Resp.FCount,
		"fType":            pid.Resp.FType,
		"qScore":           pid.Resp.QScore,
		"nmPoints":         pid.Resp.NmPoints,
		"iCount":           0,
		"iType":            0,
		"pCount":           0,
		"pType":            0,
		"srno":             srno,
		"sysid":            sysid,
		"ts":               ts,
		"pidData":          pid.Data.Data,
	}

	// ---------- Final Payload ----------
	payload := map[string]interface{}{
		"latitude":      body.Latitude,
		"longitude":     body.Longitude,
		"externalRef":   ExternalReferenceID(),
		"mobile":        body.Mobile,
		"type":          "BALANCE_ENQUIRY",
		"bankiin":       body.BankIIN,
		"biometricData": biometricData,
	}

	fmt.Println("BALANCE PAYLOAD =>", payload)

	jsonData, _ := json.Marshal(payload)

	// ---------- InstantPay API ----------
	req, _ := http.NewRequest(
		"POST",
		"https://api.instantpay.in/fi/aeps/balanceInquiry",
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
		return c.JSON(fiber.Map{
			"success": false,
			"message": "InstantPay API error",
		})
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
