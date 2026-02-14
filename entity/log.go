package entity

type Log struct {
	ID            string `json:"id"`
	EventType     string `json:"event_type"`
	ServiceName   string `json:"service_name"`
	CorrelationID string `json:"correlation_id"`
	Payload       string `json:"payload"`
	PayloadHash   string `json:"payload_hash"`
}
