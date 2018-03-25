package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	l := New()

	assert.True(t, l.Add("proj1", 1))
	assert.False(t, l.Add("proj1", 1))
	assert.True(t, l.Add("proj1", 2))
	assert.True(t, l.Add("proj1", 4))
	assert.True(t, l.Add("proj1", 3))

	assert.True(t, l.Add("proj2", 1))
}

func TestCleanup(t *testing.T) {
	l := New()

	data := []struct {
		project  string
		buildNum int
	}{
		{
			project:  "proj1",
			buildNum: 5,
		},
		{
			project:  "proj1",
			buildNum: 6,
		},
		{
			project:  "proj2",
			buildNum: 3,
		},
		{
			project:  "proj2",
			buildNum: 4,
		},
		{
			project:  "proj4",
			buildNum: 10,
		},
	}

	data2 := []struct {
		project  string
		buildNum int
	}{
		{
			project:  "later1",
			buildNum: 11,
		},
		{
			project:  "proj2",
			buildNum: 7,
		},
	}

	for _, d := range data {
		l.Add(d.project, d.buildNum)
	}

	now := time.Now()

	l.Cleanup(now.Add(-time.Hour))
	assert.Equal(t, len(data), l.size())

	l.Cleanup(now.Add(time.Hour))
	assert.Equal(t, 0, l.size())

	for _, d := range data {
		l.Add(d.project, d.buildNum)
	}

	now = time.Now()

	time.Sleep(time.Second)

	for _, d := range data2 {
		l.Add(d.project, d.buildNum)
	}

	l.Cleanup(now)
	assert.Equal(t, len(data2), l.size())
}
