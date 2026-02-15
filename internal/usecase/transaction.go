package usecase

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"transaction-service/internal/domain"
	"transaction-service/internal/infra"
)

type CreateTransactionUseCase struct {
	repo infra.TransactionRepository
}

func NewCreateTransactionUseCase(repo infra.TransactionRepository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

type CreateTransactionInput struct {
	Amount      int64
	Description string
	FromUserId  string
	ToUserId    string
}

type CreateTransactionOutput struct {
	TransactionID string
	Status        string
}

func (u *CreateTransactionUseCase) Execute(
	ctx context.Context,
	in CreateTransactionInput,
) (*CreateTransactionOutput, *infra.Exception) {
	if in.Amount <= 0 {
		return nil, infra.BadRequest(
			infra.WithCode(http.StatusBadRequest),
			infra.WithMessage("amount must be greater than zero"),
		)
	}

	if in.FromUserId == "" {
		return nil, infra.BadRequest(
			infra.WithCode(http.StatusBadRequest),
			infra.WithMessage("fromUserId is required"),
		)
	}

	if in.ToUserId == "" {
		return nil, infra.BadRequest(
			infra.WithCode(http.StatusBadRequest),
			infra.WithMessage("toUserId is required"),
		)
	}

	if in.FromUserId == in.ToUserId {
		return nil, infra.BadRequest(
			infra.WithCode(http.StatusBadRequest),
			infra.WithMessage("cannot create a transaction to the same user"),
		)
	}

	if in.Description == "" {
		return nil, infra.BadRequest(
			infra.WithCode(http.StatusBadRequest),
			infra.WithMessage("description is required"),
		)
	}

	transactionID := uuid.NewString()

	txEntity := &domain.Transaction{
		Id:                transactionID,
		Amount:            in.Amount,
		Description:       in.Description,
		FromUserId:        in.FromUserId,
		ToUserId:          in.ToUserId,
		TransactionStatus: "PENDING",
	}

	if err := u.repo.CreateTransaction(ctx, txEntity); err != nil {
		return nil, infra.Unexpected(
			infra.WithCode(http.StatusInternalServerError),
			infra.WithMessage("failed to create transaction"),
		)
	}

	return &CreateTransactionOutput{
		TransactionID: transactionID,
		Status:        string(txEntity.TransactionStatus),
	}, nil
}
