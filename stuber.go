package stuber

import (
	"github.com/bavix/features"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	MethodTitle features.Flag = iota
)

type Budgerigar struct {
	searcher *searcher
	toggles  features.Toggles
}

func NewBudgerigar(toggles features.Toggles) *Budgerigar {
	return &Budgerigar{
		searcher: newSearcher(),
		toggles:  toggles,
	}
}

func (b *Budgerigar) PutMany(values ...*Stub) []uuid.UUID {
	for _, value := range values {
		if value.Key() == uuid.Nil {
			value.ID = uuid.New()
		}
	}

	return b.searcher.upsert(values...)
}

func (b *Budgerigar) UpdateMany(values ...*Stub) []uuid.UUID {
	updates := make([]*Stub, 0, len(values))

	for _, value := range values {
		if value.Key() != uuid.Nil {
			updates = append(updates, value)
		}
	}

	return b.searcher.upsert(updates...)
}

func (b *Budgerigar) DeleteByID(ids ...uuid.UUID) int {
	return b.searcher.del(ids...)
}

func (b *Budgerigar) FindByID(id uuid.UUID) *Stub {
	return b.searcher.findByID(id)
}

func (b *Budgerigar) FindByQuery(query Query) (*Result, error) {
	// backward compatibility
	if b.toggles.Has(MethodTitle) {
		query.Method = cases.
			Title(language.English, cases.NoLower).
			String(query.Method)
	}

	return b.searcher.find(query)
}

func (b *Budgerigar) FindBy(service, method string) ([]*Stub, error) {
	return b.searcher.findBy(service, method)
}

func (b *Budgerigar) All() []*Stub {
	return b.searcher.all()
}

func (b *Budgerigar) Used() []*Stub {
	return b.searcher.used()
}

func (b *Budgerigar) Unused() []*Stub {
	return b.searcher.unused()
}

func (b *Budgerigar) Clear() {
	b.searcher.clear()
}
