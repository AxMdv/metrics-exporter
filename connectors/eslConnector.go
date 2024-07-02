package Connector

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"mertics-exporter/config"
	Data "mertics-exporter/models"
)

type ESLConnector struct {
	ESLAddress          string
	DumpPath            string
	currentLabelsStatus Data.ESLabelsPagedResult
}

func (eslConnector *ESLConnector) eslRequest(endpoint string, method string, body []byte, dump bool) []byte {
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(time.Now(), "ERROR", err)
	}
	req.Close = true
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(config.Options.Login, config.Options.Password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%v: ERROR ESL request%v\n", time.Now().Format("2006-01-02-15-04-05.000000"), err)
		}
		if len(eslConnector.DumpPath) > 0 && dump {
			dumpFileName := filepath.Join(eslConnector.DumpPath,
				fmt.Sprintf("%v_esl.json", time.Now().Format("2006-01-02-15-04-05.000000")))
			message := make(map[string]interface{})
			message["url"] = endpoint
			message["method"] = method
			message["status"] = resp.Status
			data, _ := json.Marshal(message)
			data = append(append(append([]byte("["), append(data, []byte(",")...)...), body...), []byte("]")...)
			os.WriteFile(dumpFileName, data, os.ModePerm)
		}
		return bodyBytes
	} else {
		return nil
	}
}

func (eslConnector *ESLConnector) GetServerStatus() bool {
	response := eslConnector.eslRequest("http://localhost:8001/service/status", "GET", nil, false)
	return response != nil
}

func (eslConnector *ESLConnector) GetAllLabelsStatus() []Data.ESLLabelInfo {
	url := "http://localhost:8001/service/labelinfo"
	response := eslConnector.eslRequest(url, "GET", nil, false)
	if response == nil {
		return nil
	}
	var labelsPagedResult Data.ESLabelsPagedResult
	err := json.Unmarshal(response, &labelsPagedResult)
	if err != nil {
		return make([]Data.ESLLabelInfo, 0)
	}
	eslConnector.currentLabelsStatus = labelsPagedResult
	return labelsPagedResult.LabelInfo
}

func (eslConnector *ESLConnector) GetLabelStatus(labelId string) (labelStatus Data.ESLLabelInfo) {
	url := fmt.Sprintf("http://localhost:8001/service/labelinfo/%s", labelId)
	response := eslConnector.eslRequest(url, "GET", nil, false)
	if response == nil {
		//fmt.Println("Failed to read label status")
		return
	}
	err := json.Unmarshal(response, &labelStatus)
	if err != nil {
		panic(err)
	}
	return
}

func (eslConnector *ESLConnector) FastCheckLabel(labelId string) (valid bool) {
	for _, label := range eslConnector.currentLabelsStatus.LabelInfo {
		if label.LabelId == labelId {
			valid = true
		}
	}
	return
}

func (eslConnector *ESLConnector) SendTemplateTask(templateTask Data.ESLTemplateTask) (taskId string, status bool) {
	body, _ := json.Marshal(map[string]Data.ESLTemplateTask{"TemplateTask": templateTask})
	response := eslConnector.eslRequest("http://localhost:8001/service/task", http.MethodPost, body, true)
	if response == nil {
		return
	}
	var transaction Data.ESLTransaction
	err := json.Unmarshal(response, &transaction)
	if err != nil {
		panic(err)
	}
	taskId = transaction.Id
	status = true
	return
}

func (eslConnector *ESLConnector) SendSwitchPageTask(switchPageTask Data.ESLSwitchPageTask) (taskId string, status bool) {
	body, _ := json.Marshal(map[string]Data.ESLSwitchPageTask{"SwitchPageTask": switchPageTask})
	response := eslConnector.eslRequest("http://localhost:8001/service/task", http.MethodPost, body, true)
	if response == nil {
		return
	}
	var transaction Data.ESLTransaction
	err := json.Unmarshal(response, &transaction)
	if err != nil {
		panic(err)
	}
	taskId = transaction.Id
	status = true
	return
}

func (eslConnector *ESLConnector) CheckTransactionStatus(taskId string) (finished, failed, replaced bool) {
	url := fmt.Sprintf("http://localhost:8001/service/updatestatus/transaction/%s", taskId)
	finished = true
	failed = false

	response := eslConnector.eslRequest(url, "GET", nil, false)
	if response == nil {
		//fmt.Println("Failed to read task status")
		return
	}
	var updateStatusPage Data.ESLUpdateStatusPage
	err := json.Unmarshal(response, &updateStatusPage)
	if err != nil {
		panic(err)
	}
	finished, _ = strconv.ParseBool(updateStatusPage.UpdateStatus.Finished)
	switch updateStatusPage.UpdateStatus.Status {
	case "REPLACED":
		replaced = true
	case "ERROR", "FAILED":
		failed = true
	}
	return
}

func (eslConnector *ESLConnector) CheckUpdateStatus(taskId string) (finished, failed, replaced bool) {
	url := fmt.Sprintf("http://localhost:8001/service/transaction/%s/status", taskId)

	response := eslConnector.eslRequest(url, "GET", nil, false)
	if response == nil {
		fmt.Printf("%v: ERROR Failed to read task status task: %v\n", time.Now().Format("2006-01-02-15-04-05.000000"), taskId)
		return
	}
	var transaction Data.ESLTransaction
	err := json.Unmarshal(response, &transaction)
	if err != nil {
		panic(err)
	}
	finished, _ = strconv.ParseBool(transaction.Finished)
	failed, _ = strconv.ParseBool(transaction.Failed)
	if transaction.TotalNumber == "1" && failed {
		_, _, replaced = eslConnector.CheckTransactionStatus(taskId)
	}
	return
}

func (eslConnector *ESLConnector) CancelUpdate(taskId string) bool {
	url := fmt.Sprintf("http://localhost:8001/service/transaction/%s", taskId)
	response := eslConnector.eslRequest(url, "DELETE", nil, false)
	return response != nil
}

func (eslConnector *ESLConnector) RegisterLabel(labelId string) bool {
	data := Data.ESLLabelsList{Label: append(make([]Data.ESLLabel, 0), Data.ESLLabel{Id: labelId})}
	body, _ := json.Marshal(data)
	response := eslConnector.eslRequest("http://localhost:8001/service/label", http.MethodPost, body, false)
	if response == nil {
		return false
	}
	var transaction Data.ESLTransaction
	err := json.Unmarshal(response, &transaction)
	if err != nil {
		panic(err)
	}
	if transaction.Id != "" {
		return true
	}
	return false
}

func (eslConnector *ESLConnector) UnregisterLabel(labelId string) bool {
	url := fmt.Sprintf("/service/label/%v", labelId)
	response := eslConnector.eslRequest(url, http.MethodDelete, nil, false)
	return response != nil
}

func (eslConnector *ESLConnector) EslGetServiceStatus() (totalLabels, totalAp, waitingTasks int) {
	response := eslConnector.eslRequest("http://localhost:8001/service/status", "GET", nil, false)
	if response == nil {
		fmt.Println("Failed to read service status")
		return
	}
	var eslServiceStatus Data.EslStatusData
	json.Unmarshal(response, &eslServiceStatus)
	for _, entry := range eslServiceStatus.Properties {
		if entry.Key == "waiting-tasks" {
			waitingTasks, _ = strconv.Atoi(entry.Value)
		} else if entry.Key == "labels" {
			totalLabels, _ = strconv.Atoi(entry.Value)
		} else if entry.Key == "access-points" {
			totalAp, _ = strconv.Atoi(entry.Value)
		}
	}
	return
}

func (eslConnector *ESLConnector) EslGetApStatus(totalAp int) (onlineAp int) {
	response := eslConnector.eslRequest("http://localhost:8001/service/accesspointinfo", "GET", nil, false)
	if response == nil {
		fmt.Println("Failed to read APs status")
		return
	}
	if totalAp > 1 {
		var result Data.EslApPagedResult
		json.Unmarshal(response, &result)
		for _, entry := range result.AccessPoint {
			if entry.ConnectionStatus == "ONLINE" {
				onlineAp += 1
			}
		}
	} else if totalAp == 1 {
		var result Data.EslSingleApPagedResult
		json.Unmarshal(response, &result)
		if result.AccessPoint.ConnectionStatus == "ONLINE" {
			onlineAp = 1
		}
	}
	return
}

func (eslConnector *ESLConnector) EslGetLabelsStatus() (offlineLabels, failedUpdates, failedOnlineLabels, failedGoodSignalLabels, badBatteryLabels int) {
	url := "http://localhost:8001/service/labelinfo"
	response := eslConnector.eslRequest(url, "GET", nil, false)
	if response == nil {
		fmt.Println("Failed to read labels status")
		return
	}
	var result Data.EslLabelsPagedResult
	json.Unmarshal(response, &result)
	for _, entry := range result.LabelInfo {
		if entry.ConnectionStatus == "OFFLINE" {
			offlineLabels += 1
		}
		if entry.PowerStatus == "BAD" {
			badBatteryLabels += 1
		}
		if entry.Status == "FAILED" || entry.Status == "ERROR" {
			if entry.ConnectionStatus == "ONLINE" {
				connectionStatus, _ := strconv.Atoi(entry.Rssi)
				if connectionStatus > -70 {
					failedGoodSignalLabels += 1
				}
				failedOnlineLabels += 1
			}
			failedUpdates += 1
		}
	}
	return
}
