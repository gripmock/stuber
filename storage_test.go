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

func TestLeft(t *testing.T) {
	i := newStorage()

	id := i.leftIDOrNew("Greeter")
	require.Equal(t, uint64(1), id)

	id = i.leftIDOrNew("Greeter2")
	require.Equal(t, uint64(2), id)

	id, err := i.leftID("Greeter2")
	require.NoError(t, err)
	require.Equal(t, uint64(2), id)

	id = i.leftIDOrNew("Greeter2")
	require.Equal(t, uint64(2), id)

	id, err = i.leftID("Greeter3")
	require.ErrorIs(t, ErrLeftNotFound, err)
	require.Equal(t, uint64(0), id)
}

func TestRight(t *testing.T) {
	i := newStorage()

	id := i.rightIDOrNew("SayHello")
	require.Equal(t, uint64(1), id)

	id = i.rightIDOrNew("SayHello2")
	require.Equal(t, uint64(2), id)

	id, err := i.rightID("SayHello2")
	require.NoError(t, err)
	require.Equal(t, uint64(2), id)

	id = i.rightIDOrNew("SayHello2")
	require.Equal(t, uint64(2), id)

	id, err = i.rightID("SayHello3")
	require.ErrorIs(t, ErrRightNotFound, err)
	require.Equal(t, uint64(0), id)
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

	require.Equal(t, uint64(5), s.leftTotal)
	require.Equal(t, uint64(3), s.rightTotal)
	require.Len(t, s.items, 5)
	require.Len(t, s.itemsByID, 6)
}

func TestUpdate(t *testing.T) {
	id := uuid.New()

	s := newStorage()
	s.upsert(&testItem{id: id, left: "Greeter", right: "SayHello"})

	require.Equal(t, uint64(1), s.leftTotal)
	require.Equal(t, uint64(1), s.rightTotal)
	require.Len(t, s.items, 1)
	require.Len(t, s.itemsByID, 1)

	v := s.findByID(id)
	require.NotNil(t, v)

	val, ok := v.(*testItem)
	require.True(t, ok)
	require.Equal(t, 0, val.value)

	s.upsert(&testItem{id: id, left: "Greeter", right: "SayHello", value: 42})

	require.Equal(t, uint64(1), s.leftTotal)
	require.Equal(t, uint64(1), s.rightTotal)
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

	require.Equal(t, uint64(5), s.leftTotal)
	require.Equal(t, uint64(3), s.rightTotal)
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
	require.Equal(t, uint64(3), s.leftTotal)
	require.Equal(t, uint64(3), s.rightTotal)
	require.Len(t, s.items, 3)
	require.Len(t, s.itemsByID, 3)

	require.Equal(t, 1, s.del(id1))
	require.Equal(t, uint64(3), s.leftTotal)
	require.Equal(t, uint64(3), s.rightTotal)
	require.Len(t, s.items, 2)
	require.Len(t, s.itemsByID, 2)

	require.Equal(t, 2, s.del(id2, id3))
	require.Equal(t, uint64(3), s.leftTotal)
	require.Equal(t, uint64(3), s.rightTotal)
	require.Empty(t, s.items)
	require.Empty(t, s.itemsByID)

	require.Equal(t, 0, s.del(id1, id2, id3))
	require.Equal(t, uint64(3), s.leftTotal)
	require.Equal(t, uint64(3), s.rightTotal)
	require.Empty(t, s.items)
	require.Empty(t, s.itemsByID)
}

func TestPos(t *testing.T) {
	tests := []struct {
		left  uint64
		right uint64
		guid  uuid.UUID
	}{
		{0, 0, uuid.MustParse("00000000-0000-0000-0000-000000000000")},
		{0, 1, uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		{0, 16, uuid.MustParse("00000000-0000-0000-0000-000000000010")},
		{0, 17, uuid.MustParse("00000000-0000-0000-0000-000000000011")},
		{1, 0, uuid.MustParse("00000000-0000-0001-0000-000000000000")},
		{16, 0, uuid.MustParse("00000000-0000-0010-0000-000000000000")},
		{17, 0, uuid.MustParse("00000000-0000-0011-0000-000000000000")},
		{0, 18446744073709551615, uuid.MustParse("00000000-0000-0000-ffff-ffffffffffff")},
		{18446744073709551615, 0, uuid.MustParse("ffffffff-ffff-ffff-0000-000000000000")},
		{18446744073709551615, 18446744073709551615, uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")},
	}

	for _, test := range tests {
		require.Equal(t, test.guid.String(), newStorage().pos(test.left, test.right).String())
	}
}
