package lru

import (
	"container/list"
	"errors"
)

type LRU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
}

type entry struct {
	key   interface{}
	value interface{}
}

func New(size int) (*LRU, error) {
	if size < 0 {
		return nil, errors.New("size must be positive")
	}
	lru := &LRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
	}
	return lru, nil
}

func (l *LRU) Add(key, value interface{}) bool {
	if ent, ok := l.items[key]; ok {
		l.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		return false
	}
	ent := &entry{key, value}
	entry := l.evictList.PushFront(ent)
	l.items[key] = entry
	evict := l.evictList.Len() > l.size
	if evict {
		l.RemoveOldest()
	}
	return evict
}

func (l *LRU) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := l.items[key]; ok {
		l.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return
}

func (l *LRU) GetOldest() (interface{}, interface{}, bool) {
	ent := l.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

func (l *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(l.items))
	i := 0
	for ent := l.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

func (l *LRU) Contains(key interface{}) (ok bool) {
	_, ok = l.items[key]
	return ok
}

func (l *LRU) Remove(key interface{}) bool {
	if ent, ok := l.items[key]; ok {
		l.removeElement(ent)
		return true
	}
	return false
}

func (l *LRU) RemoveOldest() (interface{}, interface{}, bool) {
	ent := l.evictList.Back()
	if ent != nil {
		l.removeElement(ent)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

func (l *LRU) removeElement(e *list.Element) {
	l.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(l.items, kv)
}

func (l *LRU) Len() int {
	return l.evictList.Len()
}

func (l *LRU) Purge() {
	for k := range l.items {
		delete(l.items, k)
	}
}
