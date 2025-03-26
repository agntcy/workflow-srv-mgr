// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

// StripedLocker implements fine-grained locking based on keys
type StripedLocker interface {
	Lock(key string)
	Unlock(key string)
}

// stripedLock is a fine-grained lock based on keys
// It is inspired from Java lib Guava: https://github.com/google/guava/wiki/StripedExplained
type stripedLock struct {
	locks []*sync.Mutex
}

func NewStripedLock(stripes uint16) StripedLocker {
	s := &stripedLock{
		locks: make([]*sync.Mutex, stripes),
	}
	for i := range s.locks {
		s.locks[i] = new(sync.Mutex)
	}
	return s
}

func (s *stripedLock) Lock(key string) {
	idx := s.keyToIndex(key)
	s.locks[idx].Lock()
}

func (s *stripedLock) Unlock(key string) {
	idx := s.keyToIndex(key)

	s.locks[idx].Unlock()
}

func (s *stripedLock) keyToIndex(key string) uint64 {
	h := xxhash.New()
	h.Write([]byte(key)) //nolint:errcheck

	idx := h.Sum64() % uint64(len(s.locks))
	return idx
}
