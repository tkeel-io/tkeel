package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
)

func TestCommonGetQueryItemsStartAndEnd(t *testing.T) {
	s, e := getQueryItemsStartAndEnd(1, 10, 0)
	assert.Equal(t, s, 0)
	assert.Equal(t, e, 0)
}

func TestSeparateEntry(t *testing.T) {
	testEntry := &v1.ConsoleEntry{
		Id: "1",
		Children: []*v1.ConsoleEntry{
			{
				Id: "1-1",
				Children: []*v1.ConsoleEntry{
					{
						Id:     "1-1-1",
						Portal: v1.ConsolePortal_tenant,
					},
					{
						Id:     "1-1-2",
						Portal: v1.ConsolePortal_tenant,
					},
				},
			},
			{
				Id:     "1-2",
				Portal: v1.ConsolePortal_tenant,
			},
			{
				Id: "1-3",
				Children: []*v1.ConsoleEntry{
					{
						Id:     "1-3-1",
						Portal: v1.ConsolePortal_admin,
					},
					{
						Id:     "1-3-2",
						Portal: v1.ConsolePortal_tenant,
					},
					{
						Id:     "1-3-3",
						Portal: v1.ConsolePortal_tenant,
					},
				},
			},
			{
				Id:     "1-4",
				Portal: v1.ConsolePortal_admin,
			},
		},
	}
	a, p := separateEntry(testEntry)
	t.Log(a)
	t.Log(p)
}

func TestMergeEntry(t *testing.T) {
	test1 := &v1.ConsoleEntry{
		Id: "1",
		Children: []*v1.ConsoleEntry{
			{
				Id: "1-1",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-1-1",
					},
					{
						Id: "1-1-2",
					},
					{
						Id: "1-1-3",
					},
				},
			},
			{
				Id: "1-2",
			},
			{
				Id: "1-3",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-3-1",
					},
					{
						Id: "1-3-2",
					},
					{
						Id: "1-3-3",
					},
				},
			},
			{
				Id: "1-4",
			},
		},
	}

	test2 := &v1.ConsoleEntry{
		Id: "1",
		Children: []*v1.ConsoleEntry{
			{
				Id: "1-1",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-1-1",
					},
					{
						Id: "1-1-2",
					},
					{
						Id: "1-1-3",
					},
				},
			},
			{
				Id: "1-2",
			},
			{
				Id: "1-3",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-3-1",
					},
					{
						Id: "1-3-2",
					},
					{
						Id: "1-3-3",
					},
				},
			},
			{
				Id: "1-4",
			},
		},
	}

	mergeEntry(test1, test2)
	t.Log(test1)
	test1 = &v1.ConsoleEntry{
		Id: "1",
		Children: []*v1.ConsoleEntry{
			{
				Id: "1-1",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-1-2",
					},
				},
			},
			{
				Id: "1-3",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-3-1",
					},
					{
						Id: "1-3-2",
					},
					{
						Id: "1-3-3",
					},
				},
			},
			{
				Id: "1-4",
			},
		},
	}
	test2 = &v1.ConsoleEntry{
		Id: "1",
		Children: []*v1.ConsoleEntry{
			{
				Id: "1-1",
				Children: []*v1.ConsoleEntry{
					{
						Id: "1-1-1",
					},
					{
						Id: "1-1-3",
					},
				},
			},
			{
				Id: "1-2",
			},
		},
	}
	mergeEntry(test1, test2)
	t.Log(test1)
}
