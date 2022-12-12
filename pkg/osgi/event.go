package osgi

type EventList struct {
	Status string  `json:"status"`
	List   []Event `json:"data"`
}

// Event represents single OSGi event
type Event struct {
	ID       string `json:"id"`
	Topic    string `json:"topic"`
	Received int64  `json:"received"`
	Category string `json:"category"`
	Info     string `json:"info"`
}

func (el *EventList) StatusUnknown() bool {
	return len(el.List) == 0
}
