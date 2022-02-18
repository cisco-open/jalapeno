// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// The contents of this file are licensed under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with the
// License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package locker

import (
	"sync"
)

// Locker defines methods to lock/unlock a key which is a string
type Locker interface {
	Lock(string)
	Unlock(string)
}

var _ Locker = &locker{}

type key struct {
	sync.Mutex
	lock *sync.Cond
	used bool
}

func newKey() *key {
	k := &key{}
	k.lock = sync.NewCond(k)
	return k
}

type locker struct {
	mtx   sync.Mutex
	store map[string]*key
}

func (l *locker) Lock(key string) {
	l.mtx.Lock()
	lock, ok := l.store[key]
	l.mtx.Unlock()
	if ok {
		lock.Lock()
		// fmt.Printf("key exists: %s used: %t\n", key, lock.used)
		// Unlocking keys map so other go routine could access it
		// Wait only if the key is already used, otherwise locked it and mark it as used.
		if lock.used {
			lock.lock.Wait()
		}
		lock.used = true
		lock.Unlock()
		return
	}
	l.mtx.Lock()
	lock = newKey()
	lock.Lock()
	lock.used = true
	l.store[key] = lock
	lock.Unlock()
	l.mtx.Unlock()
}

func (l *locker) Unlock(key string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if lock, ok := l.store[key]; ok {
		lock.Lock()
		lock.used = false
		lock.lock.Signal()
		lock.Unlock()
	}
}

// NewLocker instantiates a new instance of Locker interface
func NewLocker() Locker {
	return &locker{
		store: make(map[string]*key),
	}
}
