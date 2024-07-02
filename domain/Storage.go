package domain

import (
	"encoding/json"
	"fmt"
	Connector "mertics-exporter/connectors"
	Data "mertics-exporter/models"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Storage struct {
	TaskIdStorage IdMapStruct
	Process       ProcessStruct
	Notifications []interface{}
	BOConnector   Connector.BOConnector
	ESLConnector  Connector.ESLConnector
	Config        Data.Config
	InputQueue    chan Data.BOUpdateMessage
}

type IdMapStruct struct {
	sync.RWMutex
	IdMap map[uuid.UUID]*Entry
}

type IdMapStructItem struct {
	Key   uuid.UUID
	Value *Entry
}

func (idmap *IdMapStruct) Set(key uuid.UUID, value *Entry) {
	idmap.Lock()
	defer idmap.Unlock()
	idmap.IdMap[key] = value
}

func (idmap *IdMapStruct) Get(key uuid.UUID) (*Entry, bool) {
	idmap.RLock()
	defer idmap.RUnlock()
	value, ok := idmap.IdMap[key]
	return value, ok
}

func (idmap *IdMapStruct) Iter() <-chan IdMapStructItem {
	c := make(chan IdMapStructItem)
	f := func() {
		idmap.Lock()
		defer idmap.Unlock()

		for k, v := range idmap.IdMap {
			c <- IdMapStructItem{k, v}
		}
		close(c)
	}
	go f()
	return c
}

func (idmap *IdMapStruct) Delete(key uuid.UUID) {
	idmap.Lock()
	defer idmap.Unlock()

	delete(idmap.IdMap, key)
}

func (ps *ProcessStruct) Append(id uuid.UUID) {
	ps.Lock()
	defer ps.Unlock()
	ps.IDStorage = append(ps.IDStorage, id)
}

func (ps *ProcessStruct) Pop() (taskId uuid.UUID) {
	ps.Lock()
	defer ps.Unlock()
	taskId, ps.IDStorage = ps.IDStorage[0], ps.IDStorage[1:]
	return taskId
}

func (ps *ProcessStruct) Length() int {
	ps.RLock()
	defer ps.RUnlock()
	return len(ps.IDStorage)
}

type ProcessStruct struct {
	sync.RWMutex
	IDStorage []uuid.UUID
}

func (storage *Storage) SetReplaced(id uuid.UUID) {
	if task, success := storage.TaskIdStorage.Get(id); success {
		task.Replaced = true
	}
}

func (storage *Storage) SetComplete(id uuid.UUID) {
	if task, success := storage.TaskIdStorage.Get(id); success {
		task.Completed = true
	}
}

func (storage *Storage) GetAllTasks() (result map[uuid.UUID]*Entry) {
	storage.TaskIdStorage.RLock()
	defer storage.TaskIdStorage.RUnlock()
	result = storage.TaskIdStorage.IdMap
	return
}

func (storage *Storage) GetTaskById(id uuid.UUID) (result *Entry, flag bool) {
	if entry, ok := storage.TaskIdStorage.Get(id); ok {
		result = entry
		flag = true
	} else {
		flag = false
	}
	return
}

func (storage *Storage) GetTasksForLabel(id string) (result []*Entry) {
	for _, task := range storage.TaskIdStorage.IdMap {
		if id == task.LabelId {
			result = append(result, task)
		}
	}
	return
}

func (storage *Storage) DumpOldTasks() {
	tasksToDelete := make([]uuid.UUID, 0)
	for entry := range storage.TaskIdStorage.Iter() {
		if entry.Value.Completed && time.Now().Sub(entry.Value.Timestamp) > time.Duration(24)*time.Minute {
			data, _ := json.Marshal(entry.Value)
			dumpFileName := filepath.Join("old_api_tasks",
				fmt.Sprintf("%v_task.json", entry.Value.Timestamp.Format("2006-01-02-15-04-05.000000")))
			os.WriteFile(dumpFileName, data, os.ModePerm)
			tasksToDelete = append(tasksToDelete, entry.Key)
		}
	}
	for _, id := range tasksToDelete {
		storage.TaskIdStorage.Delete(id)
	}
}
func (storage *Storage) DumpAllTasks() {
	tasksToDelete := make([]uuid.UUID, 0)
	for entry := range storage.TaskIdStorage.Iter() {
		data, _ := json.Marshal(entry.Value)
		dumpFileName := filepath.Join("old_api_tasks",
			fmt.Sprintf("%v_task.json", entry.Value.Timestamp.Format("2006-01-02-15-04-05.000000")))
		os.WriteFile(dumpFileName, data, os.ModePerm)
		tasksToDelete = append(tasksToDelete, entry.Key)
	}
	for _, id := range tasksToDelete {
		storage.TaskIdStorage.Delete(id)
	}
}

func (storage *Storage) GetNotificationForItem(plu string) ItemNotifications {
	newNotification := ItemNotifications{Plu: plu,
		AffectedPlu:      make([]string, 0),
		SucceededLabels:  make([]*Entry, 0),
		FailedLabels:     make([]*Entry, 0),
		IncompleteLabels: make([]*Entry, 0),
		Failed:           false,
	}
	for entry := range storage.TaskIdStorage.Iter() {
		if !entry.Value.Completed {
			for _, itemId := range entry.Value.Plu {
				if itemId == plu {
					switch entry.Value.Status {
					case StatusUploaded, StatusSuccess:
						newNotification.SucceededLabels = append(newNotification.SucceededLabels, entry.Value)
					case StatusFailed:
						newNotification.FailedLabels = append(newNotification.FailedLabels, entry.Value)
					case StatusCanceled:
						break
					default:
						newNotification.IncompleteLabels = append(newNotification.IncompleteLabels, entry.Value)
					}
				} else {
					newNotification.AffectedPlu = append(newNotification.AffectedPlu, itemId)
				}
			}
		}
	}
	return newNotification
}

func (storage *Storage) GetAffectedIncompleteEntries(id uuid.UUID) []uuid.UUID {
	entry := storage.GetEntry(id)
	var affectedEntries []uuid.UUID
	var affectedPlu []string
	if entry == nil {
		return affectedEntries
	}

	affectedPlu = entry.Plu

	if len(affectedPlu) == 1 {
		plu := storage.GetAffectedPlu(affectedPlu[0])
		if plu != "" {
			affectedPlu = append(affectedPlu, plu)
		}
	}

	for _, targetPlu := range affectedPlu {
		affectedEntries = storage.GetEntryIDsForItem(targetPlu)
	}
	return affectedEntries
}

func (storage *Storage) GetTasksForPLU(plu string) (result []*Entry) {
	for task := range storage.TaskIdStorage.Iter() {
		for _, storedPlu := range task.Value.Plu {
			if plu == storedPlu {
				result = append(result, task.Value)
			}
		}
	}
	return
}

func (storage *Storage) GetEntryIDsForItem(plu string) (entries []uuid.UUID) {
	for entry := range storage.TaskIdStorage.Iter() {
		if !entry.Value.Completed {
			for _, storedPlu := range entry.Value.Plu {
				if storedPlu == plu {
					entries = append(entries, entry.Value.Id)
				}
			}
		}
	}

	return entries
}

func (storage *Storage) GetEntryStageCompletionStatus(id uuid.UUID, status string) (completed, success bool, failed []uuid.UUID) {
	originalTask, success := storage.GetTaskById(id)
	if !success {
		return false, false, make([]uuid.UUID, 0)
	}
	completed = originalTask.Status == status
	failed = make([]uuid.UUID, 0)
	if completed {
		affectedEntries := storage.GetAffectedIncompleteEntries(id)
		for _, affectedEntryId := range affectedEntries {
			task, _ := storage.TaskIdStorage.Get(affectedEntryId)

			switch task.Status {
			case StatusFailed:
				success = false
				failed = append(failed, affectedEntryId)
			case StatusUploaded:
				if status == StatusFailed {
					success = false
				}
			case StatusSuccess:
				{
				}

			default:
				return false, false, make([]uuid.UUID, 0)
			}
		}
	}
	return
}

func (storage *Storage) GetIncomplete() []uuid.UUID {
	var incompleteList []uuid.UUID
	for entry := range storage.TaskIdStorage.Iter() {
		if !entry.Value.Completed {
			incompleteList = append(incompleteList, entry.Key)
		}
	}
	return incompleteList
}

func (storage *Storage) GetCurrentGoodsNumberMetric() int {
	counter := make(map[string]int)
	for _, id := range storage.GetIncomplete() {
		task, _ := storage.TaskIdStorage.Get(id)
		for _, item := range task.Plu {
			counter[item]++
		}
	}
	return len(counter)
}

func (storage *Storage) GetActiveEntryForLabel(labelId string) (result uuid.UUID, flag bool) {
	for entry := range storage.TaskIdStorage.Iter() {
		if entry.Value.LabelId == labelId && !entry.Value.Completed {
			result = entry.Key
			flag = true
		}
	}
	return
}

func (storage *Storage) PushEntry(taskId uuid.UUID) {
	storage.Process.Append(taskId)
}

func (storage *Storage) PopEntry() (taskId uuid.UUID, valid bool) {
	if storage.Process.Length() == 0 {
		return uuid.New(), false
	}
	taskId = storage.Process.Pop()
	_, valid = storage.TaskIdStorage.Get(taskId)
	return
}

func (storage *Storage) AddTask(task *Entry) {
	storage.TaskIdStorage.Set(task.Id, task)
	storage.PushEntry(task.Id)
}

func (storage *Storage) GetEntry(id uuid.UUID) *Entry {
	entry, success := storage.TaskIdStorage.Get(id)
	if !success {
		return nil
	}
	return entry
}

func (storage *Storage) GetAffectedPlu(initialPlu string) (affectedPlu string) {
	entries := storage.GetEntryIDsForItem(initialPlu)
	for _, id := range entries {
		entry := storage.GetEntry(id)
		if len(entry.Plu) == 2 {
			for _, plu := range entry.Plu {
				if plu != initialPlu {
					affectedPlu = plu
					return
				}
			}
		}
	}
	return
}

func (storage *Storage) PushLostTasks() (flag bool) {
	for _, task := range storage.GetAllTasks() {
		if !task.Completed {
			timestamp := task.Timestamp
			for t := range task.EntryLog {
				if t.After(timestamp) {
					timestamp = t
				}
			}
			if time.Now().Sub(timestamp) > time.Duration(120) {
				storage.PushEntry(task.Id)
				flag = true
			}
		}
	}
	return
}

func (storage *Storage) ClearProcess() {
	storage.Process.Lock()
	defer storage.Process.Unlock()
	storage.Process.IDStorage = make([]uuid.UUID, 0)

}

func (storage *Storage) CountEntryInProcess() int {
	return storage.Process.Length()
}

func (storage *Storage) ProcessInputQueue() {
	articles := <-storage.InputQueue

	newTasks := make([]*Entry, 0)

	for _, article := range articles {
		// check if BOPluOrder have same EslId field

		for _, boPluOrder := range article.BOPluOrderList {
			var existingTask *Entry
			for _, task := range newTasks {
				if boPluOrder.EslId == task.LabelId {
					existingTask = task
				}
			}

			if existingTask != nil {
				if boPluOrder.PluOrder == 2 {
					existingTask.PluData = append(existingTask.PluData, article.ToEslArticle())
					existingTask.Plu = append(existingTask.Plu, article.Plu)
				} else {
					existingTask.PluData = append([]Data.CustomESLArticle{article.ToEslArticle()}, existingTask.PluData[0])
					existingTask.Plu = append([]string{article.Plu}, existingTask.Plu[0])
				}
			} else {
				task := Entry{
					LabelId:   boPluOrder.EslId,
					Id:        uuid.New(),
					Timestamp: time.Now(),
					EntryLog:  make(map[time.Time]EntryLogRecord),
				}
				task.PluData = append(task.PluData, article.ToEslArticle())
				task.Plu = []string{article.Plu}
				task.Status = StatusPopulated
				newTasks = append(newTasks, &task)
			}

		}
	}

	currentLabelsList := storage.ESLConnector.GetAllLabelsStatus()

	//var taskToReplace []uuid.UUID

	for _, task := range newTasks {
		existingEntry, entryExists := storage.GetActiveEntryForLabel(task.LabelId)
		if entryExists {
			storage.SetReplaced(existingEntry)
			//fmt.Printf("Replacing task for %v\n", storage.GetEntry(existingEntry).LabelId)
		}
		//preloadTasks.Inc()

		var labelStatus Data.ESLLabelInfo
		for _, label := range currentLabelsList {
			if label.LabelId == task.LabelId {
				labelStatus = label
			}
		}
		var err error
		task.CurrentPage, err = strconv.Atoi(labelStatus.CurrentPage)
		if err != nil {
			fmt.Printf("%v: ERROR %v label: %v\n", time.Now().Format("2006-01-02-15-04-05.000000"), err, task.LabelId)
			task.CurrentPage = 14
		}

		if labelStatus.ConnectionStatus == "ONLINE" {
			task.LabelType = labelStatus.Type
			storage.AddTask(task)
		} else if labelStatus.ConnectionStatus == "OFFLINE" {
			task.Status = StatusFailed
			storage.AddTask(task)
		}
	}
}
