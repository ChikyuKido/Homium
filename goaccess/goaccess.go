package goaccess

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"homium/utils"
	"io"
	"net/http"
)

type RequestBody struct {
	BaseURL string `json:"base_url"`
}

type Metrics struct {
	TotalRequests   int `json:"total_requests"`
	ValidRequests   int `json:"valid_requests"`
	FailedRequests  int `json:"failed_requests"`
	UniqueVisitors  int `json:"unique_visitors"`
	UniqueFiles     int `json:"unique_files"`
	Bandwidth       int `json:"bandwidth"`
	UniqueReferrers int `json:"unique_referrers"`
	UniqueNotFound  int `json:"unique_not_found"`
}

type Response struct {
	TotalRequests   int    `json:"total_requests"`
	ValidRequests   int    `json:"valid_requests"`
	FailedRequests  int    `json:"failed_requests"`
	UniqueVisitors  int    `json:"unique_visitors"`
	UniqueFiles     int    `json:"unique_files"`
	Bandwidth       string `json:"bandwidth"`
	UniqueReferrers int    `json:"unique_referrers"`
	UniqueNotFound  int    `json:"unique_not_found"`
}

type GoaccessResponse struct {
	General Metrics `json:"general"`
}

func fetchData(url string) ([]byte, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("connection error: %v", err)
	}
	defer conn.Close()

	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}

	return message, nil
}

func getStats(url string) (Response, error) {
	bytes, err := fetchData(url)
	if err != nil {
		return Response{}, fmt.Errorf("error fetching data from URL '%s': %v", url, err)
	}

	var data GoaccessResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return Response{}, fmt.Errorf("error parsing JSON data: %v", err)
	}

	return Response{
		TotalRequests:   data.General.TotalRequests,
		ValidRequests:   data.General.ValidRequests,
		FailedRequests:  data.General.FailedRequests,
		UniqueVisitors:  data.General.UniqueVisitors,
		UniqueFiles:     data.General.UniqueFiles,
		Bandwidth:       utils.ConvertMBToString(float64(data.General.Bandwidth / 1024 / 1024)),
		UniqueReferrers: data.General.UniqueReferrers,
		UniqueNotFound:  data.General.UniqueNotFound,
	}, nil
}

func GoaccessHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody RequestBody
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "unable to read request body", "details": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, `{"error": "invalid JSON format", "details": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	response, err := getStats(reqBody.BaseURL)
	if err != nil {
		http.Error(w, `{"error": "failed to get server stats", "details": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
