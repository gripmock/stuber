package stuber

type Budgerigar struct {
	storage *storage
}

func NewBudgerigar() *Budgerigar {
	return &Budgerigar{
		storage: newStorage(),
	}
}
