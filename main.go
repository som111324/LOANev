package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

const apiKey = "" // Replace with your Groq API key

type EVLoanRiskRequest struct {
	Income            int     `json:"income"`
	CreditScore       int     `json:"credit_score"`
	EmploymentYears   int     `json:"employment_years"`
	EVModel           string  `json:"ev_model"`
	BatteryHealth     float64 `json:"battery_health"`
	AnnualMileage     int     `json:"annual_mileage"`
	ChargingPattern   string  `json:"charging_pattern"`
	Location          string  `json:"location"`
	DrivingBehavior   string  `json:"driving_behavior"`
	AlternativeCredit bool    `json:"alternative_credit"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func getLoanRiskAssessment(data EVLoanRiskRequest) (string, error) {
	prompt := fmt.Sprintf(
		"Evaluate the loan risk for an EV buyer and provide a risk percentage (0-100%%) along with a brief reason. Respond in this format: { \"risk_percentage\": X, \"reason\": \"Your brief reason here.\" } Details: Income: %d, Credit Score: %d, Employment: %d years, EV Model: %s, Battery Health: %.2f%%, Annual Mileage: %d km, Charging Pattern: %s, Location: %s, Driving Behavior: %s, Alternative Credit Considered: %t.",
		data.Income, data.CreditScore, data.EmploymentYears, data.EVModel, data.BatteryHealth, data.AnnualMileage, data.ChargingPattern, data.Location, data.DrivingBehavior, data.AlternativeCredit,
	)

	payload := map[string]interface{}{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature":           1.0,
		"max_completion_tokens": 1024,
		"top_p":                 1.0,
		"stream":                false,
		"stop":                  nil,
	}

	jsonPayload, _ := json.Marshal(payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", err
	}

	if len(groqResp.Choices) > 0 {
		return groqResp.Choices[0].Message.Content, nil
	}

	return "No response from AI", nil
}

func main() {
	e := echo.New()

	e.POST("/assess-loan-risk", func(c echo.Context) error {
		var request EVLoanRiskRequest
		if err := c.Bind(&request); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		response, err := getLoanRiskAssessment(request)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"loan_risk_assessment": response})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
