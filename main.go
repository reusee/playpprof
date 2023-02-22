package main

import (
	"os"
	"runtime"
	"sync/atomic"

	"github.com/google/pprof/profile"
)

/*
ref: https://github.com/google/pprof/tree/main/proto
*/

func main() {
	p := new(profile.Profile)

	p.SampleType = append(p.SampleType, &profile.ValueType{
		Type: "cpu",
		Unit: "count",
	})

	functionIDs := make(map[uint64]bool)
	getFunction := func(function *runtime.Func, pc uintptr) *profile.Function {
		file, line := function.FileLine(pc)
		id := uint64(function.Entry())
		fn := &profile.Function{
			ID:        id,
			Name:      function.Name(),
			Filename:  file,
			StartLine: int64(line),
		}
		if _, ok := functionIDs[id]; !ok {
			p.Function = append(p.Function, fn)
			functionIDs[id] = true
		}
		return fn
	}

	var locID uint64
	getLocation := func(frame runtime.Frame) *profile.Location {
		loc := &profile.Location{
			ID: atomic.AddUint64(&locID, 1),
			Line: []profile.Line{
				{
					Function: getFunction(frame.Func, frame.PC),
					Line:     int64(frame.Line),
				},
			},
		}
		p.Location = append(p.Location, loc)
		return loc
	}

	addSample := func() {
		sample := &profile.Sample{
			Value: []int64{
				1,
			},
		}

		pcs := make([]uintptr, 1024)
		pcs = pcs[:runtime.Callers(1, pcs)]
		frames := runtime.CallersFrames(pcs)
		for {
			frame, more := frames.Next()
			if frame.Func == nil {
				// inline
				continue
			}
			location := getLocation(frame)
			sample.Location = append(sample.Location, location)

			if !more {
				break
			}
		}

		p.Sample = append(p.Sample, sample)
	}

	addSample2 := func() {
		sample := &profile.Sample{
			Value: []int64{
				1,
			},
		}

		pcs := make([]uintptr, 1024)
		pcs = pcs[:runtime.Callers(1, pcs)]
		frames := runtime.CallersFrames(pcs)
		for {
			frame, more := frames.Next()
			if frame.Func == nil {
				// inline
				continue
			}
			location := getLocation(frame)
			sample.Location = append(sample.Location, location)

			if !more {
				break
			}
		}

		p.Sample = append(p.Sample, sample)
	}

	addSample()
	addSample()
	foo(func() {
		addSample()
	})
	addSample2()

	out, err := os.Create("out")
	ce(err)
	defer out.Close()
	ce(p.Write(out))
}

//go:noline
func foo(fn func()) {
	fn()
}
