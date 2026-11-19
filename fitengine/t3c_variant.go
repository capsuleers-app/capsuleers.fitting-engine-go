package fitengine

import (
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

var t3cParentShipIDs = omap.NewSet(29984, 29986, 29988, 29990)

const parentShipAttr = 1380

var t3cGroupOrder = [4]int{958, 954, 956, 957}

// subsystemRankByTypeID e' SUBSYSTEM_RANK_BY_TYPE_ID (solo lookup).
func subsystemRankByTypeID(id int) (int, bool) {
	switch id {
	case 45626:
		return 2, true
	case 45625:
		return 3, true
	case 45627:
		return 1, true
	case 45590:
		return 2, true
	case 45589:
		return 3, true
	case 45591:
		return 1, true
	case 45603:
		return 2, true
	case 45602:
		return 3, true
	case 45601:
		return 1, true
	case 45614:
		return 1, true
	case 45615:
		return 3, true
	case 45613:
		return 2, true
	case 45624:
		return 1, true
	case 45622:
		return 2, true
	case 45623:
		return 3, true
	case 45587:
		return 1, true
	case 45586:
		return 2, true
	case 45588:
		return 3, true
	case 45600:
		return 1, true
	case 45599:
		return 2, true
	case 45598:
		return 3, true
	case 45611:
		return 1, true
	case 45612:
		return 2, true
	case 45610:
		return 3, true
	case 45629:
		return 2, true
	case 45628:
		return 3, true
	case 45630:
		return 1, true
	case 45593:
		return 2, true
	case 45592:
		return 3, true
	case 45594:
		return 1, true
	case 45605:
		return 1, true
	case 45604:
		return 2, true
	case 45606:
		return 3, true
	case 45617:
		return 3, true
	case 45616:
		return 1, true
	case 45618:
		return 2, true
	case 45632:
		return 1, true
	case 45631:
		return 3, true
	case 45633:
		return 2, true
	case 45597:
		return 3, true
	case 45596:
		return 2, true
	case 45595:
		return 1, true
	case 45608:
		return 3, true
	case 45607:
		return 2, true
	case 45609:
		return 1, true
	case 45620:
		return 1, true
	case 45619:
		return 2, true
	case 45621:
		return 3, true
	}
	return 0, false
}

// ComputeT3CVariantCode restituisce il codice di 4 cifre di un T3C completamente equipaggiato (ok=false = null).
func ComputeT3CVariantCode(shipTypeID int, subsystems []types.FitSubsystem, dataset *types.FittingDataset) (string, bool) {
	if dataset == nil {
		return "", false
	}
	if !t3cParentShipIDs.Has(shipTypeID) {
		return "", false
	}
	if len(subsystems) < len(t3cGroupOrder) {
		return "", false
	}
	subsBucket := &dataset.TypesByBucket[types.BucketSubsystems]
	fittedByGroup := omap.New[int, int]()
	for _, fs := range subsystems {
		t, ok := subsBucket.Get(fs.TypeID)
		if !ok {
			return "", false
		}
		v, has := 0.0, false
		for _, a := range t.Attributes {
			if a.ID == parentShipAttr {
				v, has = a.V, true
				break
			}
		}
		if !has || v != float64(shipTypeID) {
			return "", false
		}
		fittedByGroup.Set(t.GroupID, fs.TypeID)
	}
	code := ""
	for _, groupID := range t3cGroupOrder {
		fittedTypeID, ok := fittedByGroup.Get(groupID)
		if !ok {
			return "", false
		}
		rank, ok := subsystemRankByTypeID(fittedTypeID)
		if !ok || rank < 1 || rank > 9 {
			return "", false
		}
		code += string(rune('0' + rank))
	}
	if len(code) == len(t3cGroupOrder) {
		return code, true
	}
	return "", false
}
