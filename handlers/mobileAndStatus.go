package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mertics-exporter/domain"
	Data "mertics-exporter/models"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (h *Handlers) SendEslStatus() bool {
	eslLabels := h.Storage.ESLConnector.GetAllLabelsStatus()
	if eslLabels == nil {
		return false
	}
	boEslStatusesList := make([]Data.BOEslStatus, 0)
	for _, labelStatus := range eslLabels {
		boEslStatus := Data.BOEslStatus{EslId: labelStatus.LabelId}
		if labelStatus.PowerStatus == "GOOD" {
			boEslStatus.BatteryStatus = "OK"
		} else {
			boEslStatus.BatteryStatus = "ERROR"
		}
		if labelStatus.ConnectionStatus == "ONLINE" {
			boEslStatus.EslStatus = "OK"
		} else {
			boEslStatus.EslStatus = "ERROR"
		}
		boEslStatus.ScreenSize = Data.LabelTypesRelations[labelStatus.Type]
		boEslStatusesList = append(boEslStatusesList, boEslStatus)
	}
	boResponse := h.Storage.BOConnector.SendLabelStatus(boEslStatusesList)

	switch boResponse {
	case 204:
		return true
	case 401:
		fmt.Printf("%v: Failed to send status - Invalid token\n",
			time.Now().Format("2006-01-02-15-04-05.000000"))
		return false
	case 403:
		fmt.Printf("%v: Failed to send status - BO API is down\n",
			time.Now().Format("2006-01-02-15-04-05.000000"))
		return false
	default:
		fmt.Printf("%v: Failed to send status - Unpredicted response %v\n",
			time.Now().Format("2006-01-02-15-04-05.000000"), boResponse)
		return false
	}
}

func (h *Handlers) EslMatch(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.URL, r.Header)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, readError := io.ReadAll(r.Body)
	if readError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var matchingList []Data.Matching
	var matchingData interface{}
	var bindList []Data.BOBind

	err := json.Unmarshal(bodyBytes, &matchingData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch matchingData.(type) {
	case []interface{}:
		var matching []Data.Matching
		err := json.Unmarshal(bodyBytes, &matching)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		matchingList = append(matchingList, matching...)
	case map[string]interface{}:
		var matching Data.Matching
		err := json.Unmarshal(bodyBytes, &matching)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		matchingList = append(matchingList, matching)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, matching := range matchingList {
		bind := Data.BOBind{
			EslId: matching.LabelId,
			Plu:   matching.Plu,
		}
		bindList = append(bindList, bind)
	}

	data, boError, requestStatus := h.Storage.BOConnector.SendBind(bindList)
	var respStatus int
	var tasksList []interface{}
	fastUpdateFlag := true
	switch requestStatus {
	case 200, 207:
		newTasks := make([]*domain.Entry, 0)
		for _, article := range data {
			for _, boPluOrder := range article.BOPluOrderList {
				var existingTask *domain.Entry
				for _, task := range newTasks {
					if boPluOrder.EslId == task.LabelId {
						existingTask = task
					}
				}
				if existingTask != nil {
					existingTask.Fast = existingTask.Fast && !article.HasUpdErrors
					if boPluOrder.PluOrder == 2 {
						existingTask.PluData = append(existingTask.PluData, article.ToEslArticle())
						existingTask.Plu = append(existingTask.Plu, article.Plu)
					} else {
						existingTask.PluData = append([]Data.CustomESLArticle{article.ToEslArticle()}, existingTask.PluData[0])
						existingTask.Plu = append([]string{article.Plu}, existingTask.Plu[0])
					}
				} else {
					task := domain.Entry{
						LabelId:   boPluOrder.EslId,
						Id:        uuid.New(),
						Timestamp: time.Now(),
						EntryLog:  make(map[time.Time]domain.EntryLogRecord),
					}
					task.Fast = !article.HasUpdErrors
					task.PluData = append(task.PluData, article.ToEslArticle())
					task.Plu = []string{article.Plu}
					task.Status = domain.StatusPopulated
					if article.HasUpdErrors {
						newTasks = append(newTasks, &task)
					} else {
						for _, matching := range matchingList {
							if task.LabelId == matching.LabelId {
								newTasks = append(newTasks, &task)
							}
						}
					}
				}
			}

		}

		for _, task := range newTasks {
			MatchingTasks.Inc()
			existingEntry, entryExists := h.Storage.GetActiveEntryForLabel(task.LabelId)
			if entryExists {
				h.Storage.SetReplaced(existingEntry)
			}
			if task.Fast {
				ImageTasks.Inc()
			} else {
				PreloadTasks.Inc()
			}
			labelStatus := h.Storage.ESLConnector.GetLabelStatus(task.LabelId)
			task.CurrentPage, _ = strconv.Atoi(labelStatus.CurrentPage)
			task.LabelType = labelStatus.Type
			h.Storage.AddTask(task)
			taskStatus := map[string]interface{}{
				"taskId": task.Id.String(),
			}
			tasksList = append(tasksList, taskStatus)
			fastUpdateFlag = fastUpdateFlag && task.Fast

		}
		respStatus = http.StatusOK
	case 203:
		respStatus = http.StatusNonAuthoritativeInfo
	default:
		respStatus = http.StatusInternalServerError
	}

	w.WriteHeader(respStatus)

	responseDataBytes, _ := json.Marshal(map[string]interface{}{
		"fast":        fastUpdateFlag,
		"tasks":       tasksList,
		"boBindError": boError,
	})
	_, _ = w.Write(responseDataBytes)
	//}
}

func (h *Handlers) LabelRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, readError := io.ReadAll(r.Body)
	if readError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var matching Data.Matching
	err := json.Unmarshal(bodyBytes, &matching)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	labelIdPattern := regexp.MustCompile(`\D[\d\D]{7}`)
	if labelIdPattern.MatchString(matching.LabelId) {
		status := h.Storage.ESLConnector.RegisterLabel(matching.LabelId)
		if status {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusRequestTimeout)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

func (h *Handlers) MatchingDeletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	labelId := strings.Split(parsedUrl.Path, "/")
	labelIdPattern := regexp.MustCompile(`\D[\d\D]{7}`)
	if len(labelId) == 4 && labelIdPattern.MatchString(labelId[3]) {
		if h.BoConnection {
			switchPageTask := Data.ESLSwitchPageTask{LabelId: labelId[3], Page: "14"}
			taskId, taskUploadStatus := h.Storage.ESLConnector.SendSwitchPageTask(switchPageTask)
			if taskUploadStatus {
				entry := domain.Entry{
					LabelId:       labelId[3],
					TransactionId: taskId,
					Status:        domain.StatusSwitchProcess,
					EntryLog:      make(map[time.Time]domain.EntryLogRecord),
				}
				requestStatus := h.Storage.BOConnector.BoDeleteMatching(labelId[3])

				if requestStatus == 204 {
					h.Storage.PushEntry(entry.Id)
					w.WriteHeader(http.StatusNoContent)
				}
			}
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

func (h *Handlers) LabelDeletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	labelId := strings.Split(parsedUrl.Path, "/")
	labelIdPattern := regexp.MustCompile(`\D[\d\D]{7}`)
	if len(labelId) == 4 && labelIdPattern.MatchString(labelId[3]) {
		if h.BoConnection {
			status := h.Storage.ESLConnector.UnregisterLabel(labelId[3])
			if status {
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

//below methods for development

func (h *Handlers) ServerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	eslStatus := h.Storage.ESLConnector.GetServerStatus()
	statusMessage := Data.StatusMessage{BOStatus: h.BoConnection, ESLStatus: eslStatus}
	data, _ := json.Marshal(statusMessage)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handlers) LabelStatus(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.URL, r.Header)
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	labelId := strings.Split(parsedUrl.Path, "/")
	labelIdPattern := regexp.MustCompile(`\D[\d\D]{7}`)
	if len(labelId) == 4 && labelIdPattern.MatchString(labelId[3]) {
		labelInfo := h.Storage.ESLConnector.GetLabelStatus(labelId[3])
		if labelInfo.LabelId == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if labelInfo.ConnectionStatus == "ONLINE" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusPartialContent)
			w.Write(nil)
		}

	}
}

func (h *Handlers) GetPlu(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.URL, r.Header)
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := strings.Split(parsedUrl.Path, "/")

	pattern := regexp.MustCompile(`^\d{3,13}`)
	if len(input) == 4 && pattern.MatchString(input[3]) {
		if h.BoConnection {
			status, pluInfo := h.Storage.BOConnector.BoGetPlu(input[3])
			if status == 200 {
				data, _ := json.Marshal(pluInfo)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}
