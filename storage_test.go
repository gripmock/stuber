package stuber //nolint:testpackage

import (
	"iter"
	"maps"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testItem struct {
	id          uuid.UUID
	left, right string
	value       int
}

func (t testItem) Key() uuid.UUID {
	return t.id
}

func (t testItem) Left() string {
	return t.left
}

func (t testItem) Right() string {
	return t.right
}

func TestAdd(t *testing.T) {
	s := newStorage()
	s.upsert(
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello1"},
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello1"},
		&testItem{id: uuid.New(), left: "Greeter2", right: "SayHello2"},
		&testItem{id: uuid.New(), left: "Greeter3", right: "SayHello2"},
		&testItem{id: uuid.New(), left: "Greeter4", right: "SayHello3"},
		&testItem{id: uuid.New(), left: "Greeter5", right: "SayHello3"},
	)

	require.Len(t, s.items, 5)
	require.Len(t, s.itemsByID, 6)
}

func TestUpdate(t *testing.T) {
	id := uuid.New()

	s := newStorage()
	s.upsert(&testItem{id: id, left: "Greeter", right: "SayHello"})

	require.Len(t, s.items, 1)
	require.Len(t, s.itemsByID, 1)

	v := s.findByID(id)
	require.NotNil(t, v)

	val, ok := v.(*testItem)
	require.True(t, ok)
	require.Equal(t, 0, val.value)

	s.upsert(&testItem{id: id, left: "Greeter", right: "SayHello", value: 42})

	require.Len(t, s.items, 1)
	require.Len(t, s.itemsByID, 1)

	v = s.findByID(id)
	require.NotNil(t, v)

	val, ok = v.(*testItem)
	require.True(t, ok)
	require.Equal(t, 42, val.value)
}

func TestFindByID(t *testing.T) {
	id := uuid.MustParse("00000000-0000-0001-0000-000000000000")

	s := newStorage()
	require.Nil(t, s.findByID(id))

	s.upsert(
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello1"},
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello1"},
		&testItem{id: uuid.New(), left: "Greeter2", right: "SayHello2"},
		&testItem{id: uuid.New(), left: "Greeter3", right: "SayHello2"},
		&testItem{id: uuid.New(), left: "Greeter4", right: "SayHello3"},
		&testItem{id: uuid.New(), left: "Greeter5", right: "SayHello3"},
		&testItem{id: id, left: "Greeter1", right: "SayHello3"},
	)

	require.Len(t, s.items, 6)
	require.Len(t, s.itemsByID, 7)

	val := s.findByID(id)
	require.NotNil(t, val)
	require.Equal(t, id, val.Key())
}

func TestFindAll(t *testing.T) {
	s := newStorage()
	s.upsert(
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello1"},
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello1"},
		&testItem{id: uuid.New(), left: "Greeter2", right: "SayHello2"},
		&testItem{id: uuid.New(), left: "Greeter3", right: "SayHello2"},
		&testItem{id: uuid.New(), left: "Greeter4", right: "SayHello3"},
		&testItem{id: uuid.New(), left: "Greeter5", right: "SayHello3"},
		&testItem{id: uuid.New(), left: "Greeter1", right: "SayHello3"},
	)

	collect := func(seq iter.Seq[Value]) []Value {
		var res []Value
		for v := range seq {
			res = append(res, v)
		}

		return res
	}

	t.Run("Greeter1/SayHello1", func(t *testing.T) {
		seq, err := s.findAll("Greeter1", "SayHello1")
		require.NoError(t, err)
		require.Len(t, collect(seq), 2)
	})

	t.Run("Greeter2/SayHello2", func(t *testing.T) {
		seq, err := s.findAll("Greeter2", "SayHello2")
		require.NoError(t, err)
		require.Len(t, collect(seq), 1)
	})

	t.Run("Greeter3/SayHello2", func(t *testing.T) {
		seq, err := s.findAll("Greeter3", "SayHello2")
		require.NoError(t, err)
		require.Len(t, collect(seq), 1)
	})

	t.Run("Greeter3/SayHello3", func(t *testing.T) {
		_, err := s.findAll("Greeter3", "SayHello3")
		require.ErrorIs(t, err, ErrRightNotFound)
	})
}

func TestFindByIDs(t *testing.T) {
	s := newStorage()
	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
	s.upsert(
		&testItem{id: id1, left: "A", right: "B"},
		&testItem{id: id2, left: "C", right: "D"},
		&testItem{id: id3, left: "E", right: "F"},
	)

	t.Run("existing IDs", func(t *testing.T) {
		var results []Value
		for v := range s.findByIDs(maps.Keys(map[uuid.UUID]struct{}{id1: {}, id2: {}})) {
			results = append(results, v)
		}

		require.Len(t, results, 2)
	})

	t.Run("mixed IDs", func(t *testing.T) {
		var results []Value
		for v := range s.findByIDs(maps.Keys(map[uuid.UUID]struct{}{id1: {}, uuid.Nil: {}})) {
			results = append(results, v)
		}

		require.Len(t, results, 1)
	})
}

func TestDelete(t *testing.T) {
	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()

	s := newStorage()

	s.upsert(
		&testItem{id: id1, left: "Greeter1", right: "SayHello1"},
		&testItem{id: id2, left: "Greeter2", right: "SayHello2"},
		&testItem{id: id3, left: "Greeter3", right: "SayHello3"},
	)

	require.Equal(t, 0, s.del())
	require.Len(t, s.items, 3)
	require.Len(t, s.itemsByID, 3)

	require.Equal(t, 1, s.del(id1))
	require.Len(t, s.items, 2)
	require.Len(t, s.itemsByID, 2)

	require.Equal(t, 2, s.del(id2, id3))
	require.Empty(t, s.items)
	require.Empty(t, s.itemsByID)

	require.Equal(t, 0, s.del(id1, id2, id3))
	require.Empty(t, s.items)
	require.Empty(t, s.itemsByID)
}
