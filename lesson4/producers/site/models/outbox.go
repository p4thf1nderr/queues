package models

type Message struct {
	Payload []byte
	EventID int64 `db:"event_id"`
}

type Order struct {
	ProductID int64 `json:"product_id"`
}

type Event struct {
	ID        int64   `json:"id"`
	EventType string  `json:"event_type"`
	ProductID int64   `json:"product_id"`
	Source    string  `json:"source"`
	Timestamp float64 `json:"timestamp"`
}
