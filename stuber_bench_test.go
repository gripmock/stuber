package stuber_test

import (
	"testing"

	"github.com/bavix/features"
	"github.com/google/uuid"

	"github.com/gripmock/stuber"
)

// BenchmarkPutMany measures the performance of inserting multiple Stub values.
func BenchmarkPutMany(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Prepare a slice of Stub values to insert.
	values := make([]*stuber.Stub, 500)

	for i := range 500 {
		values[i] = &stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Insert the values into the Budgerigar.
	for range b.N {
		for range 1000 {
			budgerigar.PutMany(values...)
		}
	}
}

// BenchmarkUpdateMany measures the performance of updating multiple Stub values.
func BenchmarkUpdateMany(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values.
	values := make([]*stuber.Stub, 500)
	for i := range 500 {
		values[i] = &stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		}
	}

	budgerigar.PutMany(values...)

	// Update the values.
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		for range 1000 {
			budgerigar.UpdateMany(values...)
		}
	}
}

// BenchmarkDeleteByID measures the performance of deleting Stub values by ID.
func BenchmarkDeleteByID(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values and collect their IDs.
	ids := make([]uuid.UUID, 500)

	for i := range 500 {
		id := uuid.New()
		ids[i] = id
		budgerigar.PutMany(&stuber.Stub{
			ID:      id,
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Delete the values by their IDs.
	for _, id := range ids {
		for range 1000 {
			budgerigar.DeleteByID(id)
		}
	}
}

// BenchmarkFindByID measures the performance of finding a Stub value by ID.
func BenchmarkFindByID(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	for range 500 {
		budgerigar.PutMany(&stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Find the target value by its ID.
	for range b.N {
		for range 1000 {
			_ = budgerigar.FindByID(uuid.Nil)
		}
	}
}

// BenchmarkFindByQuery measures the performance of finding a Stub value by Query.
func BenchmarkFindByQuery(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values.
	for range 500 {
		budgerigar.PutMany(&stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	query := stuber.Query{
		Service: "service-some-name",
		Method:  "method-some-name",
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Find values by the query.
	for range b.N {
		for range 1000 {
			_, _ = budgerigar.FindByQuery(query)
		}
	}
}

// BenchmarkFindBy measures the performance of finding Stub values by service and method.
func BenchmarkFindBy(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values.
	for range 500 {
		budgerigar.PutMany(&stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	service := "service-some-name"
	method := "method-some-name"

	b.ReportAllocs()
	b.ResetTimer()

	// Find values by service and method.
	for range b.N {
		for range 1000 {
			_, _ = budgerigar.FindBy(service, method)
		}
	}
}

// BenchmarkAll measures the performance of retrieving all Stub values.
func BenchmarkAll(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values.
	for range 500 {
		budgerigar.PutMany(&stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Retrieve all values.
	for range b.N {
		for range 1000 {
			_ = budgerigar.All()
		}
	}
}

// BenchmarkUsed measures the performance of retrieving used Stub values.
func BenchmarkUsed(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values.
	for range 500 {
		budgerigar.PutMany(&stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Retrieve used values.
	for range b.N {
		for range 1000 {
			_ = budgerigar.Used()
		}
	}
}

// BenchmarkUnused measures the performance of retrieving unused Stub values.
func BenchmarkUnused(b *testing.B) {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Insert initial values.
	for range 500 {
		budgerigar.PutMany(&stuber.Stub{
			ID:      uuid.New(),
			Service: "service-" + uuid.NewString(),
			Method:  "method-" + uuid.NewString(),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	// Retrieve unused values.
	for range b.N {
		for range 1000 {
			_ = budgerigar.Unused()
		}
	}
}
