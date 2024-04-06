package stuber

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrMethodNotFound  = errors.New("method not found")
)

type Budgerigar struct {
	storage *storage
}

func NewBudgerigar() *Budgerigar {
	return &Budgerigar{
		storage: newStorage(),
	}
}

func (b *Budgerigar) PutMany(values ...*Stub) []uuid.UUID {
	return b.storage.upsert(b.castToValue(values)...)
}

func (b *Budgerigar) FindByID(id uuid.UUID) *Stub {
	if v, ok := b.storage.findByID(id).(*Stub); ok {
		return v
	}

	return nil
}

func (b *Budgerigar) DeleteByID(ids ...uuid.UUID) int {
	return b.storage.del(ids...)
}

func (b *Budgerigar) FindAll(service, method string) ([]*Stub, error) {
	all, err := b.storage.findAll(service, method)
	if err != nil {
		return nil, b.wrap(err)
	}

	return b.castToStub(all), nil
}

func (b *Budgerigar) All() []*Stub {
	return b.castToStub(b.storage.values())
}

func (b *Budgerigar) castToValue(values []*Stub) []Value {
	result := make([]Value, len(values))
	for i, v := range values {
		result[i] = v
	}

	return result
}

func (b *Budgerigar) castToStub(values []Value) []*Stub {
	ret := make([]*Stub, 0, len(values))
	for _, v := range values {
		if s, ok := v.(*Stub); ok {
			ret = append(ret, s)
		}
	}

	return ret
}

func (b *Budgerigar) wrap(err error) error {
	if errors.Is(err, ErrLeftNotFound) {
		return ErrServiceNotFound
	}

	if errors.Is(err, ErrRightNotFound) {
		return ErrMethodNotFound
	}

	return err
}
