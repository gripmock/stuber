package stuber

import (
	"github.com/bavix/features"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MethodTitle is a feature flag for using title casing in the method field
// of a Query struct.
const MethodTitle features.Flag = iota

// Budgerigar is the main struct for the stuber package. It contains a
// searcher and toggles.
type Budgerigar struct {
	searcher *searcher
	toggles  features.Toggles
}

// NewBudgerigar creates a new Budgerigar with the given features.Toggles.
//
// Parameters:
// - toggles: The features.Toggles to use.
//
// Returns:
// - A new Budgerigar.
func NewBudgerigar(toggles features.Toggles) *Budgerigar {
	return &Budgerigar{
		searcher: newSearcher(),
		toggles:  toggles,
	}
}

// PutMany inserts the given Stub values into the Budgerigar. If a Stub value
// does not have a key, a new UUID is generated for its key.
//
// Parameters:
// - values: The Stub values to insert.
//
// Returns:
// - []uuid.UUID: The keys of the inserted Stub values.
func (b *Budgerigar) PutMany(values ...*Stub) []uuid.UUID {
	// Iterate over each Stub value.
	for _, value := range values {
		// If the Stub value does not have a key, generate a new UUID for its key.
		if value.Key() == uuid.Nil {
			value.ID = uuid.New()
		}
	}

	// Insert the Stub values into the Budgerigar's searcher.
	return b.searcher.upsert(values...)
}

func (b *Budgerigar) UpdateMany(values ...*Stub) []uuid.UUID {
	// Extract the values that have a non-nil key.
	// These values will be updated in the searcher.
	updates := make([]*Stub, 0, len(values))

	for _, value := range values {
		// Only update the value if it has a non-nil key.
		if value.Key() != uuid.Nil {
			updates = append(updates, value)
		}
	}

	// Insert the updates into the searcher.
	// Returns the keys of the inserted or updated values.
	//
	// Parameters:
	// - values: The Stub values to insert or update.
	//
	// Returns:
	// - []uuid.UUID: The keys of the inserted or updated values.
	return b.searcher.upsert(updates...)
}

// DeleteByID deletes the Stub values with the given IDs from the Budgerigar's searcher.
//
// Parameters:
// - ids: The UUIDs of the Stub values to delete.
//
// Returns:
// - int: The number of Stub values that were successfully deleted.
func (b *Budgerigar) DeleteByID(ids ...uuid.UUID) int {
	// Delete the Stub values with the given IDs from the searcher.
	// Returns the number of Stub values that were successfully deleted.
	//
	// Parameters:
	// - ids: The UUIDs of the Stub values to delete.
	//
	// Returns:
	// - int: The number of Stub values that were successfully deleted.
	return b.searcher.del(ids...)
}

// FindByID retrieves the Stub value associated with the given ID from the Budgerigar's searcher.
//
// Parameters:
// - id: The UUID of the Stub value to retrieve.
//
// Returns:
// - *Stub: The Stub value associated with the given ID, or nil if not found.
func (b *Budgerigar) FindByID(id uuid.UUID) *Stub {
	// FindByID retrieves the Stub value associated with the given ID from the Budgerigar's searcher.
	//
	// Parameters:
	// - id: The UUID of the Stub value to retrieve.
	//
	// Returns:
	// - *Stub: The Stub value associated with the given ID, or nil if not found.
	return b.searcher.findByID(id)
}

// FindByQuery retrieves the Stub value associated with the given Query from the Budgerigar's searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (b *Budgerigar) FindByQuery(query Query) (*Result, error) {
	// Backward compatibility: convert the method field to title case if the MethodTitle feature flag is enabled.
	if b.toggles.Has(MethodTitle) {
		query.Method = cases.
			Title(language.English, cases.NoLower).
			String(query.Method)
	}

	// Find the Stub value associated with the given Query from the Budgerigar's searcher.
	//
	// Parameters:
	// - query: The Query used to search for a Stub value.
	//
	// Returns:
	// - *Result: The Result containing the found Stub value (if any), or nil.
	// - error: An error if the search fails.
	return b.searcher.find(query)
}

// FindBy retrieves all Stub values that match the given service and method
// from the Budgerigar's searcher.
//
// Parameters:
// - service: The service field used to search for Stub values.
// - method: The method field used to search for Stub values.
//
// Returns:
// - []*Stub: The Stub values that match the given service and method, or nil if not found.
// - error: An error if the search fails.
func (b *Budgerigar) FindBy(service, method string) ([]*Stub, error) {
	return b.searcher.findBy(service, method)
}

// All returns all Stub values from the Budgerigar's searcher.
//
// Returns:
// - []*Stub: All Stub values.
func (b *Budgerigar) All() []*Stub {
	return b.searcher.all()
}

// Used returns all Stub values that have been used from the Budgerigar's searcher.
//
// Returns:
// - []*Stub: All used Stub values.
func (b *Budgerigar) Used() []*Stub {
	return b.searcher.used()
}

// Unused returns all Stub values that have not been used from the Budgerigar's searcher.
//
// Returns:
// - []*Stub: All unused Stub values.
func (b *Budgerigar) Unused() []*Stub {
	return b.searcher.unused()
}

// Clear clears all Stub values from the Budgerigar's searcher.
func (b *Budgerigar) Clear() {
	b.searcher.clear()
}
