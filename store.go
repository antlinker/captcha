// Copyright 2011 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"container/list"
	"sync"
	"time"
)

// Store An object implementing Store interface can be registered with SetCustomStore
// function to handle storage and retrieval of captcha ids and solutions for
// them, replacing the default memory store.
//
// It is the responsibility of an object to delete expired and used captchas
// when necessary (for example, the default memory store collects them in Set
// method after the certain amount of captchas has been stored.)
type Store interface {
	// Set sets the digits for the captcha id.
	Set(id string, digits []byte, maxcheckCnt int)

	// Get returns stored digits for the captcha id. Clear indicates
	// whether the captcha must be deleted from the store.
	Get(id string, clear bool) (digits []byte, maxCheckCnt int)
	GetForID(id string) (digits []byte, maxCheckCnt int, checkCnt int)
}

// expValue stores timestamp and id of captchas. It is used in the list inside
// memoryStore for indexing generated captchas by timestamp to enable garbage
// collection of expired captchas.
type idByTimeValue struct {
	timestamp time.Time
	id        string
}
type idValue struct {
	value       []byte
	checkCnt    int
	maxCheckCnt int
}

// memoryStore is an internal store for captcha ids and their values.
type memoryStore struct {
	sync.RWMutex
	digitsByID map[string]*idValue
	idByTime   *list.List
	// Number of items stored since last collection.
	numStored int
	// Number of saved items that triggers collection.
	collectNum int
	// Expiration time of captchas.
	expiration time.Duration
}

// NewMemoryStore returns a new standard memory store for captchas with the
// given collection threshold and expiration time (duration). The returned
// store must be registered with SetCustomStore to replace the default one.
func NewMemoryStore(collectNum int, expiration time.Duration) Store {
	s := new(memoryStore)
	s.digitsByID = make(map[string]*idValue)
	s.idByTime = list.New()
	s.collectNum = collectNum
	s.expiration = expiration
	return s
}

func (s *memoryStore) Set(id string, digits []byte, maxcheckCnt int) {
	s.Lock()
	s.digitsByID[id] = &idValue{value: digits, maxCheckCnt: maxcheckCnt}
	s.idByTime.PushBack(idByTimeValue{time.Now(), id})
	s.numStored++
	if s.numStored <= s.collectNum {
		s.Unlock()
		return
	}
	s.Unlock()
	go s.collect()
}
func (s *memoryStore) GetForID(id string) (digits []byte, maxCheckCnt int, checkCnt int) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.digitsByID[id]
	if !ok {
		return
	}
	return val.value, val.maxCheckCnt, val.checkCnt
}
func (s *memoryStore) Get(id string, clear bool) (digits []byte, maxCheckCnt int) {
	if !clear {
		// When we don't need to clear captcha, acquire read lock.
		s.RLock()
		defer s.RUnlock()
	} else {
		s.Lock()
		defer s.Unlock()
	}
	val, ok := s.digitsByID[id]
	if !ok {
		return
	}
	maxCheckCnt = val.maxCheckCnt
	if !clear && val.checkCnt < val.maxCheckCnt {
		digits = val.value
		return
	}
	val.checkCnt++
	if val.checkCnt > val.maxCheckCnt {
		// 已经验证到指定次数
		return
	}
	digits = val.value
	if val.checkCnt >= val.maxCheckCnt {
		delete(s.digitsByID, id)
		// XXX(dchest) Index (s.idByTime) will be cleaned when
		// collecting expired captchas.  Can't clean it here, because
		// we don't store reference to expValue in the map.
		// Maybe store it?
	}
	return
}

func (s *memoryStore) collect() {
	now := time.Now()
	s.Lock()
	defer s.Unlock()
	s.numStored = 0
	for e := s.idByTime.Front(); e != nil; {
		ev, ok := e.Value.(idByTimeValue)
		if !ok {
			return
		}
		if ev.timestamp.Add(s.expiration).Before(now) {
			delete(s.digitsByID, ev.id)
			next := e.Next()
			s.idByTime.Remove(e)
			e = next
		} else {
			return
		}
	}
}
