package domain

type ItemNotifications struct {
	Plu              string
	AffectedPlu      []string
	SucceededLabels  []*Entry
	FailedLabels     []*Entry
	IncompleteLabels []*Entry
	Failed           bool
}

func (notification *ItemNotifications) create(plu string) {
	notification.Plu = plu
}

func (notification *ItemNotifications) getNotification() interface{} {
	if len(notification.IncompleteLabels) == 0 {
		if len(append(notification.SucceededLabels, notification.FailedLabels...)) == 0 {
			return nil
		}
		updateStatus := map[string]interface{}{}
		updateStatus["plu"] = notification.Plu
		if len(notification.FailedLabels) == 0 {
			if notification.Failed {
				updateStatus["updStatus"] = "ERROR"
			} else {
				updateStatus["updStatus"] = "OK"
			}
		} else {
			updateStatus["updStatus"] = "ERROR"
			var labelsList []string
			for _, entry := range notification.FailedLabels {
				labelsList = append(labelsList, entry.LabelId)
			}
			updateStatus["eslID"] = labelsList
			for _, entry := range notification.SucceededLabels {
				entry.Completed = true
			}
		}
		return updateStatus
	} else {
		return nil
	}

}

func printLabelsList(input []*Entry) (result []string) {
	for _, entry := range input {
		result = append(result, entry.LabelId)
	}
	return
}
