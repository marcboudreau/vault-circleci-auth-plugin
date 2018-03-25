package main

import (
	"container/heap"
	"time"
)

var exists = struct{}{}

// CircleCIBuildList is a list that holds build identifiers sorted by time.
type CircleCIBuildList struct {
	projects   map[string]map[int]struct{}
	buildTimes BuildHeap
}

// New ...
func New() *CircleCIBuildList {
	return &CircleCIBuildList{
		projects:   make(map[string]map[int]struct{}),
		buildTimes: BuildHeap{},
	}
}

// Add attempts to add a build record and returns true if there was no matching record,
// otherwise false is returned.
func (p *CircleCIBuildList) Add(project string, buildNum int) bool {
	if _, ok := p.projects[project]; !ok {
		p.projects[project] = make(map[int]struct{})
	}

	if _, ok := p.projects[project][buildNum]; ok {
		return false
	}

	p.projects[project][buildNum] = exists

	build := CircleCIBuild{
		time:     time.Now(),
		project:  project,
		buildNum: buildNum,
	}

	heap.Push(&p.buildTimes, build)

	return true
}

// Cleanup removes all build records recorded prior to t.
func (p *CircleCIBuildList) Cleanup(t time.Time, b *backend) {
	if b != nil {
		b.Logger().Trace("cleaning up attempts made before", t.Format(time.UnixDate))
	}

	if len(p.buildTimes) > 0 {
		build := heap.Pop(&p.buildTimes).(CircleCIBuild)

		for build.time.Before(t) {
			if _, ok := p.projects[build.project]; ok {
				if b != nil {
					b.Logger().Trace("  removing", build.project, build.buildNum, build.time.Format(time.UnixDate))
				}
				delete(p.projects[build.project], build.buildNum)

				if len(p.projects[build.project]) == 0 {
					delete(p.projects, build.project)
				}
			}

			if len(p.buildTimes) > 0 {
				build = heap.Pop(&p.buildTimes).(CircleCIBuild)
			} else {
				return
			}
		}

		heap.Push(&p.buildTimes, build)
	}
}

func (p *CircleCIBuildList) size() int {
	return p.buildTimes.Len()
}

// BuildHeap ...
type BuildHeap []CircleCIBuild

// CircleCIBuild ...
type CircleCIBuild struct {
	time     time.Time
	project  string
	buildNum int
}

func (h BuildHeap) Less(i, j int) bool {
	return h[i].time.Before(h[j].time)
}

func (h BuildHeap) Len() int {
	return len(h)
}

func (h BuildHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push ...
func (h *BuildHeap) Push(x interface{}) {
	*h = append(*h, x.(CircleCIBuild))
}

// Pop ...
func (h *BuildHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
