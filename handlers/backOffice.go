package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	Data "mertics-exporter/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *Handlers) verifyToken(r *http.Request) (valid bool, err error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		err := fmt.Errorf("invalid bearer format")
		return false, err
	}
	reqToken = strings.TrimSpace(splitToken[1])
	if reqToken != h.Storage.BOConnector.BOToken {
		return false, nil
	} else {
		return true, nil
	}

}

func (h *Handlers) BoUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//check if token is valid
	valid, err := h.verifyToken(r)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		fmt.Println("Token validation completed")
	}
	var articles Data.BOUpdateMessage
	bodyBytes, readError := io.ReadAll(r.Body)
	if readError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(bodyBytes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &articles)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var unregisteredLabelsList []string
	if len(h.Storage.Config.DumpFlag) > 0 {
		dumpFileName := filepath.Join(h.Storage.Config.DumpFlag,
			fmt.Sprintf("%v_gkupdate.json", time.Now().Format("2006-01-02-15-04-05.000000")))
		_ = os.WriteFile(dumpFileName, bodyBytes, os.ModePerm)
	}

	for _, article := range articles {
		for _, label := range article.BOPluOrderList {
			if !h.Storage.ESLConnector.FastCheckLabel(label.EslId) {
				unregisteredLabelsList = append(unregisteredLabelsList, label.EslId)
			}
		}
	}

	if len(unregisteredLabelsList) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		data, _ := json.Marshal(unregisteredLabelsList)
		w.WriteHeader(http.StatusPartialContent)
		w.Write(data)
	}

	h.Storage.InputQueue <- articles
}
