package domain

import (
	"fmt"
	Data "mertics-exporter/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func Process(taskId uuid.UUID, storage *Storage) {
	entry, _ := storage.GetTaskById(taskId)
	if storage.Config.Debug {
		fmt.Println(time.Now().Format("2006-01-02-15-04-05.000000"), taskId, entry.Status, entry.LabelId, entry.Plu)
	}
	if !entry.Completed && !entry.Replaced {
		switch status := entry.Status; status {
		case StatusPopulated:
			eslTaskId, successFlag := storage.ESLConnector.SendTemplateTask(CreateTemplateTask(entry))
			if successFlag {
				entry.TransactionId = eslTaskId
				if entry.Fast {
					entry.Status = StatusSwitchProcess
				} else {
					entry.Status = StatusUploadProcess
				}
				entry.EntryLog[time.Now()] = EntryLogRecord{Status: entry.Status, Message: ""}
			} else {
				entry.Status = StatusFailed
				entry.EntryLog[time.Now()] = EntryLogRecord{
					Status:  entry.Status,
					Message: "Failed to send template task",
				}
			}
		case StatusUploadProcess:
			finished, failed, replaced := storage.ESLConnector.CheckUpdateStatus(entry.TransactionId)
			if finished {
				if failed {
					entry.Status = StatusFailed
					if replaced {
						entry.Replaced = true
						entry.Completed = true
						entry.EntryLog[time.Now()] = EntryLogRecord{
							Status:  entry.Status,
							Message: "Upload task replaced",
						}
					} else {
						entry.EntryLog[time.Now()] = EntryLogRecord{
							Status:  entry.Status,
							Message: "Upload task failed",
						}
					}
				} else {
					entry.Status = StatusUploaded
					entry.EntryLog[time.Now()] = EntryLogRecord{Status: entry.Status, Message: ""}
				}
			}
		case StatusUploaded:
			complete, success, failed := storage.GetEntryStageCompletionStatus(entry.Id, StatusUploaded)

			if complete {
				for _, plu := range entry.Plu {
					notification := storage.GetNotificationForItem(plu)
					for _, task := range notification.SucceededLabels {
						if success {
							task.Status = StatusUploadForArticle
							task.EntryLog[time.Now()] = EntryLogRecord{
								Status:  entry.Status,
								Message: fmt.Sprintf("Upload task success for: %s", printLabelsList(notification.SucceededLabels)),
							}
						} else {
							if task.Status != StatusFailed {
								task.Completed = true
								task.Status = StatusCanceled
								task.EntryLog[time.Now()] = EntryLogRecord{
									Status:  StatusCanceled,
									Message: fmt.Sprintf("Canseled due to fail on: %v", failed),
								}
							}
						}
					}
				}
			}
		case StatusUploadForArticle:
			eslTaskId, successFlag := storage.ESLConnector.SendSwitchPageTask(CreateSwitchPageTask(entry))
			if successFlag {
				entry.TransactionId = eslTaskId
				entry.Status = StatusSwitchProcess
				entry.EntryLog[time.Now()] = EntryLogRecord{Status: entry.Status, Message: ""}

			} else {
				entry.Status = StatusFailed
				entry.EntryLog[time.Now()] = EntryLogRecord{
					Status:  entry.Status,
					Message: "Failed to send switch page task",
				}
			}
		case StatusSwitchProcess:
			finished, failed, replaced := storage.ESLConnector.CheckUpdateStatus(entry.TransactionId)
			if finished {
				if failed {
					entry.Status = StatusFailed
					if replaced {
						entry.Completed = true
						entry.Replaced = true
					}
					entry.EntryLog[time.Now()] = EntryLogRecord{
						Status:  entry.Status,
						Message: "Failed to switch page",
					}

				} else {
					entry.Status = StatusSuccess
					entry.EntryLog[time.Now()] = EntryLogRecord{Status: entry.Status, Message: ""}
				}
			}

		case StatusRejected:
			entry.Completed = true

		case StatusFailed, StatusSuccess:
			complete, _, _ := storage.GetEntryStageCompletionStatus(entry.Id, entry.Status)
			if complete {
				for _, itemId := range entry.Plu {
					notification := storage.GetNotificationForItem(itemId)
					message := notification.getNotification()
					if message != nil {
						storage.Notifications = append(storage.Notifications, message)
					}
				}
				for _, affectedEntry := range storage.GetAffectedIncompleteEntries(entry.Id) {
					storage.SetComplete(affectedEntry)
				}
			}
		}
	} else if entry.Replaced {
		storage.ESLConnector.CancelUpdate(entry.TransactionId)
		entry.Completed = true
	}

}

func CreateTemplateTask(entry *Entry) (templateTask Data.ESLTemplateTask) {
	if entry.CurrentPage == 1 {
		templateTask.Page = "2"
	} else {
		templateTask.Page = "1"
	}
	templateTask.LabelId = entry.LabelId
	templateTask.Preload = fmt.Sprint(!entry.Fast)
	templateTask.SkipOnEqualImage = "false"
	if len(entry.Plu) == 2 {
		if strings.HasSuffix(entry.LabelType, "BWY") {
			templateTask.Template = "api_double_bwy.xsl"
		} else {
			templateTask.Template = "api_double_bwr.xsl"
		}
	} else {
		if strings.HasSuffix(entry.LabelType, "BWY") {
			templateTask.Template = "api_default_bwy.xsl"
		} else {
			templateTask.Template = "api_default_bwr.xsl"
		}
	}

	templateTask.Article = map[string]Data.CustomESLArticle{"article1": entry.PluData[0]}
	if len(entry.PluData) == 2 {
		templateTask.Article["article2"] = entry.PluData[1]
	}
	templateTask.TaskPriority = "NORMAL"
	return
}

func CreateSwitchPageTask(entry *Entry) (switchPageTask Data.ESLSwitchPageTask) {
	if entry.CurrentPage == 1 {
		switchPageTask.Page = "2"
	} else {
		switchPageTask.Page = "1"
	}
	switchPageTask.LabelId = entry.LabelId
	return
}
