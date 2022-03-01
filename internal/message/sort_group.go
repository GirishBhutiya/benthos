package message

import (
	"context"
)

// SortGroup associates a tag of a part with the original group.
type SortGroup struct {
	Len int
}

// NewSortGroupParts creates a sort group associated with a slice of parts.
func NewSortGroupParts(parts []*Part) (*SortGroup, []*Part) {
	g := &SortGroup{Len: len(parts)}
	newParts := make([]*Part, len(parts))

	for i, part := range parts {
		tag := &tag{
			Index: i,
			Group: g,
		}

		ctx := GetContext(part)

		var prev tagChecker
		if v, ok := ctx.Value(tagKey).(tagChecker); ok {
			prev = v
		}

		ctx = context.WithValue(ctx, tagKey, tagValue{
			tag:      tag,
			previous: prev,
		})

		newParts[i] = WithContext(ctx, part)
	}

	return g, newParts
}

// NewSortGroup creates a new sort group to be associated with a
func NewSortGroup(m *Batch) (*SortGroup, *Batch) {
	inParts := make([]*Part, m.Len())
	_ = m.Iter(func(i int, part *Part) error {
		inParts[i] = part
		return nil
	})

	group, outParts := NewSortGroupParts(inParts)
	newMsg := QuickBatch(nil)
	newMsg.SetAll(outParts)

	return group, newMsg
}

// GetIndex attempts to determine the original index of a message part relative
// to a sort group.
func (g *SortGroup) GetIndex(p *Part) int {
	ctx := GetContext(p)

	v, ok := ctx.Value(tagKey).(tagChecker)
	if !ok {
		return -1
	}

	return v.IndexForGroup(g)
}

//------------------------------------------------------------------------------

type tag struct {
	Index int
	Group groupType
}

type tagType *tag

type groupType *SortGroup

type tagKeyType int

const tagKey tagKeyType = iota

type tagChecker interface {
	IndexForGroup(g groupType) int
	HasTag(t tagType) bool
}

type tagValue struct {
	tag      tagType
	previous tagChecker
}

func (t tagValue) IndexForGroup(g groupType) int {
	if t.tag.Group == g {
		return t.tag.Index
	}
	if t.previous != nil {
		return t.previous.IndexForGroup(g)
	}
	return -1
}

func (t tagValue) HasTag(tag tagType) bool {
	if t.tag == tag {
		return true
	}
	if t.previous != nil {
		return t.previous.HasTag(tag)
	}
	return false
}
