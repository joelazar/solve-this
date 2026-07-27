package domain

import (
	"sort"
	"time"
)

var Now = time.Now

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

var priorityRank = map[Priority]int{
	PriorityHigh:   0,
	PriorityMedium: 1,
	PriorityLow:    2,
}

func (p Priority) Rank() int {
	if rank, ok := priorityRank[p]; ok {
		return rank
	}
	return len(priorityRank)
}

func (p Priority) Valid() bool {
	_, ok := priorityRank[p]
	return ok
}

type List struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type Task struct {
	ID        string
	ListID    string
	Title     string
	Notes     string
	Done      bool
	Priority  Priority
	Tags      map[string]struct{}
	Due       *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Task) Touch() {
	t.UpdatedAt = Now()
}

func (t Task) Clone() Task {
	clone := t
	if t.Tags != nil {
		clone.Tags = make(map[string]struct{}, len(t.Tags))
		for tag := range t.Tags {
			clone.Tags[tag] = struct{}{}
		}
	}
	if t.Due != nil {
		due := *t.Due
		clone.Due = &due
	}
	return clone
}

func (t Task) Overdue(now time.Time) bool {
	return !t.Done && t.Due != nil && t.Due.Before(now)
}

func (t Task) HasTag(tag string) bool {
	_, ok := t.Tags[tag]
	return ok
}

func (t Task) TagList() []string {
	tags := make([]string, 0, len(t.Tags))
	for tag := range t.Tags {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func (t *Task) AddTag(tag string) {
	if t.Tags == nil {
		t.Tags = make(map[string]struct{})
	}
	t.Tags[tag] = struct{}{}
}

func (t *Task) RemoveTag(tag string) bool {
	if _, ok := t.Tags[tag]; !ok {
		return false
	}
	delete(t.Tags, tag)
	return true
}

func (t *Task) SetTags(tags []string) {
	t.Tags = make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		t.Tags[tag] = struct{}{}
	}
}
