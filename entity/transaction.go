package entity

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "PENDING"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCompleted  TransactionStatus = "COMPLETED"
	StatusFailed     TransactionStatus = "FAILED"
)

type Transaction struct {
	Id                string            `json:"id"`
	Amount            int64             `json:"amount"`
	Description       string            `json:"description"`
	FromUserId        string            `json:"fromUserId"`
	ToUserId          string            `json:"toUserId"`
	TransactionStatus TransactionStatus `json:"transactionStatus"`
	CreatedAt         string            `json:"createdAt"`
	ProcessedAt       string            `json:"processedAt"`
}
