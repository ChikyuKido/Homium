package crafty

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"homium/utils"
	"io"
	"net/http"
)

type RequestBody struct {
	APIToken string `json:"api_token"`
	BaseURL  string `json:"base_url"`
}

type CraftyResponse struct {
	Status string     `json:"status"`
	Data   ServerStat `json:"data"`
}

type ServerStat struct {
	Mem       string  `json:"mem"`
	WorldSize string  `json:"world_size"`
	Online    int     `json:"online"`
	CPU       float32 `json:"cpu"`
	Running   bool    `json:"running"`
	Crashed   bool    `json:"crashed"`
	Max       int     `json:"max"`
}

type Response struct {
	Mem           string  `json:"mem"`
	WorldSize     string  `json:"world_size"`
	Players       int     `json:"players"`
	MaxPlayers    int     `json:"max_players"`
	CPU           float32 `json:"cpu"`
	ServerRunning int     `json:"server_running"`
	ServerCrashed int     `json:"server_crashed"`
	ServerTotal   int     `json:"server_total"`
	ServerOffline int     `json:"server_offline"`
}

func aggregateStats(apiToken, baseURL string) (Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", baseURL+"servers", nil)
	if err != nil {
		return Response{}, fmt.Errorf("error creating request for servers endpoint: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("error sending request to servers endpoint: %v", err)
	}
	defer resp.Body.Close()

	var serversData struct {
		Data []struct {
			ServerID string `json:"server_id"`
		} `json:"data"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("error reading response body: %v", err)
	}
	err = json.Unmarshal(body, &serversData)
	if err != nil {
		return Response{}, fmt.Errorf("error parsing JSON response: %v", err)
	}

	var accMem, accWorldSize float64
	var accPlayers, accServerRunning, accServerCrashed, accMaxPlayers int
	var accCPU float32
	accServerTotal := len(serversData.Data)

	for _, server := range serversData.Data {
		req, err := http.NewRequest("GET", baseURL+"servers/"+server.ServerID+"/stats", nil)
		if err != nil {
			return Response{}, fmt.Errorf("error creating request for server stats endpoint: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiToken)
		resp, err := client.Do(req)
		if err != nil {
			return Response{}, fmt.Errorf("error sending request to server stats endpoint: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return Response{}, fmt.Errorf("error reading response body for server stats: %v", err)
		}
		var craftyResponse CraftyResponse
		err = json.Unmarshal(body, &craftyResponse)
		if err != nil {
			return Response{}, fmt.Errorf("error parsing JSON response for server stats: %v", err)
		}

		serverStat := craftyResponse.Data
		if serverStat.Mem != "0" {
			mem, _ := utils.ConvertToMB(serverStat.Mem)
			worldSize, _ := utils.ConvertToMB(serverStat.WorldSize)

			accMem += mem
			accWorldSize += worldSize
			accPlayers += serverStat.Online
			accCPU += serverStat.CPU
			accMaxPlayers += serverStat.Max

			if serverStat.Running {
				accServerRunning++
			}
			if serverStat.Crashed {
				accServerCrashed++
			}
		}
	}

	accServerOffline := accServerTotal - accServerRunning

	return Response{
		Mem:           utils.ConvertMBToString(accMem),
		WorldSize:     utils.ConvertMBToString(accWorldSize),
		Players:       accPlayers,
		MaxPlayers:    accMaxPlayers,
		CPU:           accCPU,
		ServerRunning: accServerRunning,
		ServerCrashed: accServerCrashed,
		ServerTotal:   accServerTotal,
		ServerOffline: accServerOffline,
	}, nil
}

func CraftyHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := aggregateStats(reqBody.APIToken, reqBody.BaseURL+"/api/v2/")
	if err != nil {
		http.Error(w, `{"error": "failed to get server stats", "details": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
