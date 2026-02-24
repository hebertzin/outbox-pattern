package usecase

type (
	createTransactionFeature interface {
		CreateTransaction() error
	}

	CreateTransactionUseCase struct {
		feature createTransactionFeature
	}
)

func NewCreateTransaction(feature createTransactionFeature) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{feature: feature}
}

func (uc *CreateTransactionUseCase) CreateTransaction() error {
	err := uc.feature.CreateTransaction()
	if err != nil {
		return err
	}

	return nil
}
