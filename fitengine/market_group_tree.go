package fitengine

import (
	"fmt"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// MarketGroupPlacement e' MarketGroupPlacement.
type MarketGroupPlacement struct {
	CategoryKey, CategoryName, SubGroupKey, SubGroupName string
}

// MarketGroupPlacementFor colloca un type nell'albero Market a due livelli (ricade su categoria/gruppo SDE
// se il bundle non ha market group).
func MarketGroupPlacementFor(t *types.SdeType, dataset *types.FittingDataset) MarketGroupPlacement {
	if dataset.MarketGroups.Len() == 0 {
		return sdeFallback(t, dataset)
	}
	type node struct {
		id   int
		name *string
	}
	var chain []node
	cursor := t.MarketGroupID
	seen := omap.NewSet[int]()
	for cursor != nil && !seen.Has(*cursor) {
		seen.Add(*cursor)
		n, ok := dataset.MarketGroups.Get(*cursor)
		if !ok {
			break
		}
		chain = append(chain, node{n.ID, n.Name})
		cursor = n.ParentGroupID
	}
	if len(chain) == 0 {
		return sdeFallback(t, dataset)
	}
	leaf := chain[0]
	parent := leaf
	if len(chain) > 1 {
		parent = chain[1]
	}
	nameOr := func(n node) string {
		if n.name != nil {
			return *n.name
		}
		return fmt.Sprintf("Market group %d", n.id)
	}
	return MarketGroupPlacement{
		CategoryKey: fmt.Sprintf("mg-%d", parent.id), CategoryName: nameOr(parent),
		SubGroupKey: fmt.Sprintf("mg-%d", leaf.id), SubGroupName: nameOr(leaf),
	}
}

func sdeFallback(t *types.SdeType, dataset *types.FittingDataset) MarketGroupPlacement {
	name := fmt.Sprintf("Group %d", t.GroupID)
	if g, ok := dataset.Groups.Get(t.GroupID); ok && g.Name != nil {
		name = *g.Name
	}
	return MarketGroupPlacement{CategoryKey: "cat-other", CategoryName: "Other", SubGroupKey: fmt.Sprintf("group-%d", t.GroupID), SubGroupName: name}
}
