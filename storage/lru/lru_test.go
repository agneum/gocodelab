package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRU_add(t *testing.T) {
	l, err := New(1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if l.Add(1, 1) == true {
		t.Errorf("should not have an eviction")
	}

	if l.Add(2, 2) == false {
		t.Errorf("should have an eviction")
	}
}

func TestLRU_contains(t *testing.T) {
	l, err := New(2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	l.Add(1, 1)
	l.Add(2, 2)

	if !l.Contains(1) {
		t.Errorf("1 should be contained")
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("contains should not have updated recent-ness of 1")
	}
}

func TestLRU_GetOldest_RemoveOldest(t *testing.T) {
	l, err := New(128)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}

	k, _, ok := l.GetOldest()
	if !ok {
		t.Fatalf("missing element")
	}
	if k.(int) != 128 {
		t.Fatalf("bad value: %v", k)
	}

	k, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing element")
	}
	if k.(int) != 128 {
		t.Fatalf("bad value: %v", k)
	}

	k, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing element")
	}
	if k.(int) != 129 {
		t.Fatalf("wrong value: %v", k)
	}
}

func TestLRU(t *testing.T) {
	l, err := New(128)
	assert.NoError(t, err)

	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	assert.Equal(t, 128, l.Len())

	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || k != v || v != i+128 {
			t.Fatalf("wrong key: %v", k)
		}
	}

	for i := 0; i < 128; i++ {
		_, ok := l.Get(i)
		assert.False(t, ok)
	}
	for i := 128; i < 256; i++ {
		_, ok := l.Get(i)
		assert.True(t, ok)
	}
	for i := 128; i < 192; i++ {
		ok := l.Remove(i)
		assert.True(t, ok)
		ok = l.Remove(i)
		assert.False(t, ok)
		_, ok = l.Get(i)
		assert.False(t, ok)
	}

	l.Get(192)

	for i, k := range l.Keys() {
		if (i < 63 && i != i+193) || (i == 63 && k != 192) {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		t.Fatalf("wrong len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		t.Fatalf("should contain nothing")
	}
}
