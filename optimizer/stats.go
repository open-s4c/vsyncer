// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"fmt"
	"log"
	"math"
	"time"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Type

// Type represents the measurement type, eg, Success of NotSafe.
type Type int

const (
	// Success count
	Success Type = iota
	// NotSafe count
	NotSafe
	// NotLive count
	NotLive
	// Skip optimization count
	Skip
	// Invalid count: could not translate (eg, ACQ in write)
	Invalid
	// Ignore count: should ignore (masks for example)
	Ignore
	// Error count
	Error
	// Total number of tests
	Total
	// Timeout considered to be OK
	Timeout
)

type timeStats struct {
	sum  float64
	sum2 float64
	cnt  int
}

// Stats keeps tracks of count stats and timing measurements
type Stats struct {
	counts map[Type]int
	start  time.Time
	first  time.Time
	last   time.Time
	time   map[string]timeStats
}

// NewStats returns a new Stats object
func NewStats() *Stats {
	return &Stats{
		counts: make(map[Type]int),
		start:  time.Now(),
		time:   make(map[string]timeStats),
	}
}

// Inc increments the stats count of type t
func (s *Stats) Inc(t Type) {
	if t == Success {
		s.last = time.Now()
		if s.counts[Success] == 0 {
			s.first = s.last
		}
	}
	s.counts[t]++
}

// AddTime adds a time durations to a tag
func (s *Stats) AddTime(tag string, d time.Duration) {
	t := s.time[tag]
	t.sum += float64(d)
	t.sum2 += float64(d) * float64(d)
	t.cnt++
	s.time[tag] = t
}

func (ts timeStats) mean() time.Duration {
	sum := float64(ts.sum)
	cnt := float64(ts.cnt)
	mean := time.Duration(sum / cnt)
	return mean
}

func (ts timeStats) sd() time.Duration {
	sum := float64(ts.sum)
	cnt := float64(ts.cnt)

	sum2 := float64(ts.sum2)
	sd := time.Duration(math.Sqrt(sum2/cnt - math.Pow(sum/cnt, u2)))
	return sd
}

// String is the string representation of the stats object.
func (s *Stats) String() string {
	var str string
	for k, v := range s.counts {
		str += fmt.Sprintf("%8v: %d\n", k, v)
	}

	elapsed := time.Since(s.start)
	str += fmt.Sprintf("\nTotal time: %v (%v)\n", elapsed.Seconds(), elapsed)

	for tag, tstats := range s.time {
		str += fmt.Sprintf("Mean time %s: %v (sd=%v cnt=%v)\n", tag, tstats.mean(), tstats.sd(), tstats.cnt)
	}
	return str
}

// GetTime returns the time duration spent with a specific tag.
func (s *Stats) GetTime(tag string) (time.Duration, time.Duration) {
	log.Println(tag, "------")

	if tstats, has := s.time[tag]; has {
		log.Printf("Mean time %s: %v (sd=%v cnt=%v)\n", tag, tstats.mean(), tstats.sd(), tstats.cnt)
		return tstats.mean(), tstats.sd()
	}
	return 0, 0
}
