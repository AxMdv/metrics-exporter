package domain

import (
	Data "mertics-exporter/models"
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	Id            uuid.UUID
	Timestamp     time.Time
	TransactionId string
	Plu           []string
	LabelId       string
	Status        string
	CurrentPage   int
	Fast          bool
	Completed     bool
	Replaced      bool
	PluData       []Data.CustomESLArticle
	LabelType     string
	EntryLog      map[time.Time]EntryLogRecord
}

type EntryLogRecord struct {
	Status  string
	Message string
}
