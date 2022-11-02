// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package goid

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"unsafe"
)

func TestTypeGoID(t *testing.T) {
	var gid GoID = 4711
	var gidIfc interface{} = gid

	// GoID is a strong type and should not allow type assertion to int64
	if _, ok := gidIfc.(int64); ok {
		t.Errorf("type assertion from GoID to int64 succeeded, should not")
	}

	// Make sure reflect recognizes type GoID as an int64. Many marshalling
	// libraries will depend on that.
	gidVal := reflect.ValueOf(gidIfc)
	gidType := gidVal.Type()
	if gidType.Kind() != reflect.Int64 {
		t.Errorf("expected gidType.Kind() to be Int64, got %s", gidType.Kind().String())
	}

	// Make sure GoID is formatted as an int64
	if s := fmt.Sprint(gid); s != "4711" {
		t.Errorf("fmt.Sprint(gid) printed %q", s)
	}
	if s := fmt.Sprintf("%v", gid); s != "4711" {
		t.Errorf(`fmt.Sprintf("%%v", gid) printed %q`, s)
	}
	if s := fmt.Sprintf("%#v", gid); s != "4711" {
		t.Errorf(`fmt.Sprintf("%%#v", gid) printed %q`, s)
	}
	if s := fmt.Sprintf("%d", gid); s != "4711" {
		t.Errorf(`fmt.Sprintf("%%d", gid) printed %q`, s)
	}
}

func TestGetGidOffset(t *testing.T) {
	if getGidOffset() < 0 {
		t.Fatalf("getGidOffset failed unexpectedly")
	}

	// let slowGid() fail
	temp := goroutinePrefix
	defer func() {
		goroutinePrefix = temp
	}()
	goroutinePrefix = "fake "
	if getGidOffset() >= 0 {
		t.Fatalf("getGidOffset succeeded unexpectedly")
	}
}

func TestFindGidOffset(t *testing.T) {
	if off := findGidOffset(10, 9); off >= 0 {
		t.Errorf("expected findGidOffset(%d,%d) to find nothing, found offset %d", 10, 9, off)
	}
	if off := findGidOffset(0, gSize); off < 0 {
		t.Errorf("findGidOffset(%d,%d) failed to find anything", 0, gSize)
	}

	var foundCnt int
	for off := 0; ; {
		off = findGidOffset(off, gSize)
		if off != -1 {
			foundCnt++
			off += (int)(unsafe.Sizeof(GoID(0)))
		} else {
			break
		}
	}
	if foundCnt == 0 {
		t.Fatal("findGidOffset failed to find anything")
	}
}

func testGid(t *testing.T, getGid func() GoID) {
	t.Helper()
	gidMap := make(map[GoID]bool)
	waitCh := make(chan bool)
	testCount := 1000
	mu := &sync.Mutex{}
	for i := 0; i < testCount; i++ {
		go func() {
			mu.Lock()
			defer mu.Unlock()
			gid := fastGid()
			gidMap[gid] = true
			if gid > 0 {
				waitCh <- true
			} else {
				waitCh <- false
			}
		}()
	}
	for i := 0; i < testCount; i++ {
		if !<-waitCh {
			t.Fatalf("zero gid found")
		}
	}
	if len(gidMap) != testCount {
		t.Fatalf("duplicate gid found")
	}
}

func TestFastGid(t *testing.T) {
	testGid(t, fastGid)
}

func TestSlowGid(t *testing.T) {
	testGid(t, slowGid)
}

func TestGetGoID(t *testing.T) {
	// fastGid
	testGid(t, GetGoID)

	// slowGid
	temp := gidOffset
	defer func() {
		gidOffset = temp
	}()
	gidOffset = -1
	testGid(t, GetGoID)
}

// To disable dead code optimization which would defeat the benchmarks
var Unused GoID

func BenchmarkSlowGid(b *testing.B) {
	b.ReportAllocs()
	var gid GoID
	for i := 0; i < b.N; i++ {
		gid = slowGid()
	}
	Unused = gid
}

func BenchmarkFastGid(b *testing.B) {
	b.ReportAllocs()
	var gid GoID
	for i := 0; i < b.N; i++ {
		gid = fastGid()
	}
	Unused = gid
}

func BenchmarkGetGoID(b *testing.B) {
	b.ReportAllocs()
	var gid GoID
	for i := 0; i < b.N; i++ {
		gid = GetGoID()
	}
	Unused = gid
}
