package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func Estamp(c *fiber.Ctx) error {

	// -------- Request Body Struct --------
	type Signature struct {
		Height   int      `json:"height"`
		OnPages  []string `json:"onPages"`
		Position string   `json:"position"`
		Width    int      `json:"width"`
	}

	type Signer struct {
		Identifier  string    `json:"identifier"`
		DisplayName string    `json:"displayName"`
		BirthYear   string    `json:"birthYear"`
		Signature   Signature `json:"signature"`
	}

	type EstampRequest struct {
		DocumentID          string   `json:"documentId"`
		RedirectURL         string   `json:"redirectUrl"`
		EstampState         int      `json:"estampState"`
		EstampValue         int      `json:"estampValue"`
		EstampMergePosition int      `json:"estampMergePosition"`
		Reason              string   `json:"reason"`
		Signers             []Signer `json:"signers"`
	}

	// -------- Parse Incoming Request --------
	var req EstampRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// -------- Convert to JSON --------
	jsonData, err := json.Marshal(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to marshal request",
		})
	}

	// -------- Create HTTP Request --------
	apiURL := "https://dg-sandbox.setu.co/api/signature/estamp"

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create request",
		})
	}

	// -------- Headers --------
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-client-id", "0b4649a2-0f7d-4d10-b015-9e38c6d3a799")
	httpReq.Header.Set("x-client-secret", "Ib404T5CtvKrXKmpzMs8sQgxphMwg9bW")
	httpReq.Header.Set("x-product-instance-id", "b41b4ca5-4037-4bb3-97b3-b23b52e6237a")

	// -------- Call API --------
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "API call failed",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	// -------- Read Response --------
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to read response",
		})
	}
	// ✅ Log Response
	fmt.Println("===== API Response =====")
	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))
	// -------- Return Response --------
	return c.Status(resp.StatusCode).Send(body)
}
