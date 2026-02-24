package usecase

import (
	"context"
	"net/http"
	"transaction-service/internal/errors"

	"github.com/google/uuid"

	"transaction-service/internal/domain"
	"transaction-service/internal/infra"
)

type (
	CreateTransactionExecutor interface {
		Execute(ctx context.Context, in CreateTransactionInput) (*CreateTransactionOutput, *errors.Exception)
	}

	CreateTransactionUseCase struct {
		repo infra.TransactionRepository
	}

	CreateTransactionInput struct {
		Amount      int64
		Description string
		FromUserId  string
		ToUserId    string
	}

	CreateTransactionOutput struct {
		TransactionID string
		Status        string
	}
)

func NewCreateTransactionUseCase(repo infra.TransactionRepository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

func (u *CreateTransactionUseCase) Execute(
	ctx context.Context,
	in CreateTransactionInput,
) (*CreateTransactionOutput, *errors.Exception) {
	if in.Amount <= 0 {
		return nil, errors.BadRequest(
			errors.WithCode(http.StatusBadRequest),
			errors.WithMessage("amount must be greater than zero"),
		)
	}

	if in.FromUserId == "" {
		return nil, errors.BadRequest(
			errors.WithCode(http.StatusBadRequest),
			errors.WithMessage("fromUserId is required"),
		)
	}

	if in.ToUserId == "" {
		return nil, errors.BadRequest(
			errors.WithCode(http.StatusBadRequest),
			errors.WithMessage("toUserId is required"),
		)
	}

	if in.FromUserId == in.ToUserId {
		return nil, errors.BadRequest(
			errors.WithCode(http.StatusBadRequest),
			errors.WithMessage("cannot create a transaction to the same user"),
		)
	}

	if in.Description == "" {
		return nil, errors.BadRequest(
			errors.WithCode(http.StatusBadRequest),
			errors.WithMessage("description is required"),
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
		return nil, errors.Unexpected(
			errors.WithCode(http.StatusInternalServerError),
			errors.WithMessage("failed to create transaction"),
		)
	}

	return &CreateTransactionOutput{
		TransactionID: transactionID,
		Status:        string(txEntity.TransactionStatus),
	}, nil
}
