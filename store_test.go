// Copyright 2011 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"bytes"
	"testing"
)

func TestSetGet(t *testing.T) {
	s := NewMemoryStore(CollectNum, Expiration)
	id := "captcha id"
	d := RandomDigits(10)
	max := 2
	s.Set(id, d, max)
	d2, qmax := s.Get(id, false)
	if d2 == nil || !bytes.Equal(d, d2) || max != qmax {
		t.Errorf("saved %v(%d), getDigits returned got %v(%d)", d, max, d2, qmax)
	}
}

func TestGetClear(t *testing.T) {
	s := NewMemoryStore(CollectNum, Expiration)
	id := "captcha id"
	d := RandomDigits(10)
	max := 1
	s.Set(id, d, max)
	d2, _ := s.Get(id, true)
	if d2 == nil || !bytes.Equal(d, d2) {
		t.Errorf("saved %v, getDigitsClear returned got %v", d, d2)
	}
	d2, _ = s.Get(id, false)
	if d2 != nil {
		t.Errorf("getDigitClear didn't clear (%q=%v)", id, d2)
	}
}
func TestGetClearForNum(t *testing.T) {
	s := NewMemoryStore(CollectNum, Expiration)
	id := "captcha id"
	d := RandomDigits(10)
	max := 5
	s.Set(id, d, max)
	for i := 0; i < max; i++ {

		d2, _ := s.Get(id, true)
		if d2 == nil || !bytes.Equal(d, d2) {
			t.Errorf("saved %v, getDigitsClear returned got %v", d, d2)
		}
	}
	d2, _ := s.Get(id, true)
	if d2 != nil {
		t.Errorf("getDigitClear didn't clear (%q=%v)", id, d2)
	}
	d2, _ = s.Get(id, false)
	if d2 != nil {
		t.Errorf("getDigitClear didn't clear (%q=%v)", id, d2)
	}
}
func TestCollect(t *testing.T) {
	//TODO(dchest): can't test automatic collection when saving, because
	//it's currently launched in a different goroutine.
	s := NewMemoryStore(10, -1)
	// create 10 ids
	ids := make([]string, 10)
	d := RandomDigits(10)
	for i := range ids {
		ids[i] = randomId()
		s.Set(ids[i], d, 1)
	}
	s.(*memoryStore).collect()
	// Must be already collected
	nc := 0
	for i := range ids {
		d2, _ := s.Get(ids[i], false)
		if d2 != nil {
			t.Errorf("%d: not collected", i)
			nc++
		}
	}
	if nc > 0 {
		t.Errorf("= not collected %d out of %d captchas", nc, len(ids))
	}
}

func BenchmarkSetCollect(b *testing.B) {
	b.StopTimer()
	d := RandomDigits(10)
	s := NewMemoryStore(9999, -1)
	ids := make([]string, 1000)
	for i := range ids {
		ids[i] = randomId()
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			s.Set(ids[j], d, 1)
		}
		s.(*memoryStore).collect()
	}
}
