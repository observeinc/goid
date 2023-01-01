package goid

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"unsafe"
)

// GoID is a goroutine id, a 64-bit integer that identifies a goroutine
type GoID int64

// GetGoID gets the current goroutine id
func GetGoID() GoID {
	if FastGetGoIDAvailable() {
		return fastGid()
	}
	return slowGid()
}

// FastGetGoIDAvailable tells if a fast way to get current goroutine id is
// available. GetGoID will use a very slow path otherwise
func FastGetGoIDAvailable() bool {
	return gidOffset >= 0
}

// getg returns the "g", a control block that holds runtime information about
// the current goroutine. Implemented in Assembly.
//
//go:noescape
func getg() *g

// Just for type safety. The contents of the "g" are only known to package
// runtime and may change between Go versions.
type g struct{}

var (
	goroutinePrefix = "goroutine "
	gidOffset       = getGidOffset() // Runs once during package initialization
)

const (
	gidSize    = (int)(unsafe.Sizeof(GoID(0)))
	gSize      = 256 // If this library ever breaks, try to up this constant
	checkCount = 10  // Number of checks per candidate offset, by each voter
	voterCount = 10
)

// slowGid calls runtime.Stack and extracts the goroutine id from the
// stacktrace
func slowGid() GoID {
	buf := [32]byte{}

	// Parse the 4707 out of "goroutine 4707 ["
	str := strings.TrimPrefix(
		string(buf[:runtime.Stack(buf[:], false)]),
		goroutinePrefix,
	)

	if lastOffset := strings.IndexByte(str, ' '); lastOffset > 0 {
		if id, err := strconv.ParseInt(str[:lastOffset], 10, gidSize*8); err == nil {
			return GoID(id)
		}
	}
	return 0
}

// fastGid extracts the goroutine id from the "g"
func fastGid() GoID {
	return gidFromG(getg(), gidOffset)
}

// gidFromG casts the value at `g + offset` to a GoID
//
//go:nocheckptr
func gidFromG(g *g, offset int) GoID {
	return *(*GoID)(unsafe.Pointer(uintptr(unsafe.Pointer(g)) + uintptr(offset)))
}

// findGidOffset iterates from `getg() + startOffset` to `getg() + maxOffset`
// and returns the first offset where the stored value matches slowGid()
func findGidOffset(startOffset, maxOffset int) (offset int) {
	currGid := slowGid()
	g := getg()

	// Handle segmentation faults in case we run past the "g"
	oldPanicOnFault := debug.SetPanicOnFault(true)
	defer func() {
		if r := recover(); r != nil {
			offset = -1
		}
	}()
	defer func() { debug.SetPanicOnFault(oldPanicOnFault) }()

	if currGid != 0 && g != nil {
		for offset = startOffset; offset < maxOffset; offset += gidSize {
			if gidFromG(g, offset) == currGid {
				return offset
			}
		}
	}
	return -1
}

// checkGidOffset spawns a bunch of goroutines and tests whether the value
// stored at `getg() + offset` matches what is returned by slowGid(). Returns
// true if and only if the value matches for all spawned goroutines.
func checkGidOffset(offset int) bool {
	ret := make(chan bool, checkCount)

	for i := 0; i < checkCount; i++ {
		go func() {
			gid := slowGid()
			g := getg()
			defer func() {
				if r := recover(); r != nil {
					ret <- false
				}
			}()
			match := gid != 0 &&
				g != nil &&
				gidFromG(g, offset) == gid
			ret <- match
		}()
	}

	result := true
	for i := 0; i < checkCount; i++ {
		if !<-ret {
			result = false
		}
	}
	return result
}

// getGidOffset figures out the offset in the "g" where the goroutine id is
// stored
func getGidOffset() int {
	// Spawn a bunch of "voter" goroutines, each of which finds a set of
	// candidate offsets which appear to contain goroutine ids according
	// to checkGidOffset
	ret := make(chan []int, voterCount)
	for i := 0; i < voterCount; i++ {
		go func() {
			var localCandidateOffsets []int
			for offset := 0; offset < gSize; offset += gidSize {
				offset = findGidOffset(offset, gSize)
				if offset == -1 {
					// No more candidate offsets past offset
					break
				}
				if checkGidOffset(offset) {
					localCandidateOffsets = append(localCandidateOffsets, offset)
				}
			}
			ret <- localCandidateOffsets
		}()
	}

	// Count the votes
	globalCandidateOffsets := make(map[int]int)
	for i := 0; i < voterCount; i++ {
		for _, offset := range <-ret {
			globalCandidateOffsets[offset]++
		}
	}

	// Pick an offset which all voters agree on. It is overwhelmingly likely
	// that it is truly a valid offset where "g" stores the goroutine id.
	for offset, votes := range globalCandidateOffsets {
		if votes == voterCount {
			return offset
		}
	}

	// No such offset found
	return -1
}
