package Connector

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	Data "mertics-exporter/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/99designs/keyring"
)

type BOConnector struct {
	BOAddress string
	DumpPath  string
	BOToken   string
}

func (boConnector *BOConnector) boRequest(endpoint, method string, body []byte) ([]byte, int) {
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Print(err)
	}
	req.Close = true
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", boConnector.BOToken))
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusBadRequest
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNonAuthoritativeInfo {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Print(err)
		}
		if len(boConnector.DumpPath) > 0 {
			dumpFileName := filepath.Join(boConnector.DumpPath,
				fmt.Sprintf("%v_bo.json", time.Now().Format("2006-01-02-15-04-05.000000")))
			message := make(map[string]interface{})
			message["url"] = endpoint
			message["method"] = method
			message["status"] = resp.Status
			data, _ := json.Marshal(message)
			data = append(append(append([]byte("["), append(data, []byte(",")...)...), body...), []byte("]")...)
			os.WriteFile(dumpFileName, data, os.ModePerm)

		}
		return bodyBytes, resp.StatusCode
	} else {
		if len(boConnector.DumpPath) > 0 {
			dumpFileName := filepath.Join(boConnector.DumpPath,
				fmt.Sprintf("%v_bo.json", time.Now().Format("2006-01-02-15-04-05.000000")))
			message := make(map[string]interface{})
			message["url"] = endpoint
			message["method"] = method
			message["status"] = resp.Status
			data, _ := json.Marshal(message)
			data = append(append(append([]byte("["), append(data, []byte(",")...)...), body...), []byte("]")...)
			os.WriteFile(dumpFileName, data, os.ModePerm)
		}
		return nil, resp.StatusCode
	}
}

func (boConnector *BOConnector) SendLabelStatus(status []Data.BOEslStatus) (requestStatus int) {
	body, _ := json.Marshal(status)
	url := fmt.Sprintf("https://%v/esl/status", boConnector.BOToken)
	_, requestStatus = boConnector.boRequest(url, http.MethodPut, body)
	return
}

func (boConnector *BOConnector) SendUpdateStatus(updateStatus []interface{}) (requestStatus int) {
	url := fmt.Sprintf("https://%v/esl/updstatus", boConnector.BOToken)
	body, _ := json.Marshal(updateStatus)
	_, requestStatus = boConnector.boRequest(url, http.MethodPut, body)
	return
}

func (boConnector *BOConnector) SendBind(boBindList []Data.BOBind) (boGoods []Data.BOArticle, boError []Data.BOErrorDescription, requestStatus int) {
	url := fmt.Sprintf("https://%v/esl/bind", boConnector.BOToken)
	body, _ := json.Marshal(boBindList)
	data, requestStatus := boConnector.boRequest(url, http.MethodPut, body)
	switch requestStatus {
	case 200:
		json.Unmarshal(data, &boGoods)
	case 203:
		json.Unmarshal(data, &boError)
	}
	return
}

func (boConnector *BOConnector) BoDeleteMatching(labelId string) (requestStatus int) {
	url := fmt.Sprintf("https://%v/esl/unbind/%v", boConnector.BOToken, labelId)
	_, requestStatus = boConnector.boRequest(url, http.MethodDelete, nil)
	return
}

func (boConnector *BOConnector) BoGetPlu(input string) (requestStatus int, pluInfo Data.PluInfo) {
	url := fmt.Sprintf("https://%v/esl/pluInfo/%v", boConnector.BOToken, input)
	data, requestStatus := boConnector.boRequest(url, http.MethodGet, nil)
	json.Unmarshal(data, &pluInfo)
	return
}

func (boConnector *BOConnector) BoDeleteLabel(labelId string) (requestStatus int) {
	url := fmt.Sprintf("https://%v/esl/delete/%v", boConnector.BOToken, labelId)
	_, requestStatus = boConnector.boRequest(url, http.MethodDelete, nil)
	return
}

func (boConnector *BOConnector) BoGetToken() (err error) {
	ring, err := keyring.Open(keyring.Config{
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		ServiceName:      "smartapi",
		FileDir:          "./connectors",
		FilePasswordFunc: keyring.FixedStringPrompt("smartapi"),
	})
	if err != nil {
		return err
	}
	i, err := ring.Get("jwt-bo")
	if err != nil {
		return err
	}
	boConnector.BOToken = string(i.Data)
	return err
}
