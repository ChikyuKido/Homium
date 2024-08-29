package goaccess

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"homium/utils"
	"io"
	"log"
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
		log.Fatal("Error while connecting to the WebSocket server:", err)
	}
	defer conn.Close()

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Error while reading message:", err)
	}

	return message, nil
}

func getStats(url string) (Response, error) {
	bytes, err := fetchData(url)
	if err != nil {
		return Response{}, fmt.Errorf("failed to fetch HTML: %v", err)
	}

	var data GoaccessResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return Response{}, err
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
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := getStats(reqBody.BaseURL)
	if err != nil {
		http.Error(w, "Failed to get server stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
