package stuber //nolint:testpackage

import (
	"iter"
	"testing"

	"github.com/google/uuid"
)

//nolint:gochecknoinits
func init() {
	uuid.EnableRandPool()
}

func BenchmarkStorageValues(b *testing.B) {
	items := make([]Value, 0, b.N)
	for range b.N {
		items = append(items, &testItem{id: uuid.New(), left: "A", right: "B"})
	}

	s := newStorage()
	s.upsert(items...)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		for range s.values() { //nolint:revive
		}
	}
}

func BenchmarkStorageFindAll(b *testing.B) {
	items := make([]Value, 0, b.N)
	for range b.N {
		items = append(items, &testItem{id: uuid.New(), left: "A", right: "B"})
	}

	s := newStorage()
	s.upsert(items...)

	var all iter.Seq[Value]

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		all, _ = s.findAll("A", "B")
		for range all { //nolint:revive
		}
	}
}

func BenchmarkStorageFindByID(b *testing.B) {
	items := make([]Value, 0, b.N)
	for range b.N {
		items = append(items, &testItem{id: uuid.New(), left: "A", right: "B"})
	}

	s := newStorage()
	s.upsert(items...)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = s.findByID(uuid.New())
	}
}

func BenchmarkStorageDel(b *testing.B) {
	items := make([]Value, 0, b.N)
	for range b.N {
		items = append(items, &testItem{id: uuid.New(), left: "A", right: "B"})
	}

	s := newStorage()
	s.upsert(items...)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = s.del(uuid.New())
	}
}

func BenchmarkStoragePosByN(b *testing.B) {
	s := newStorage()
	s.upsert(&testItem{id: uuid.New(), left: "A", right: "B"})

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = s.posByN("A", "B")
	}
}

func BenchmarkStoragePos(b *testing.B) {
	s := newStorage()

	left := s.leftIDOrNew("A")
	right := s.rightIDOrNew("B")

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = s.pos(left, right)
	}
}

func BenchmarkStorageLeftID(b *testing.B) {
	s := newStorage()
	s.upsert(&testItem{id: uuid.New(), left: "A", right: "B"})

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = s.leftID("A")
	}
}

func BenchmarkStorageLeftIDOrNew(b *testing.B) {
	s := newStorage()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = s.leftIDOrNew(uuid.NewString())
	}
}

func BenchmarkStorageRightID(b *testing.B) {
	s := newStorage()
	s.upsert(&testItem{id: uuid.New(), left: "A", right: "B"})

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = s.rightID("B")
	}
}

func BenchmarkStorageRightIDOrNew(b *testing.B) {
	s := newStorage()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = s.rightIDOrNew(uuid.NewString())
	}
}
