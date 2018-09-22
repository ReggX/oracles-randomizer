package main

import (
	"container/list"
	"math/rand"
	"sort"

	"github.com/jangler/oos-randomizer/graph"
	"github.com/jangler/oos-randomizer/logic"
	"github.com/jangler/oos-randomizer/rom"
)

// returns true iff the node is in the list.
func nodeInList(n *graph.Node, l *list.List) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(*graph.Node) == n {
			return true
		}
	}
	return false
}

func trySlotRandomItem(r *Route, src *rand.Rand,
	itemPool, slotPool *list.List, countFunc func(r *Route) int,
	numUsedSlots int, fillUnused bool) (usedItem, usedSlot *list.Element) {
	// we're dead
	if slotPool.Len() == 0 || itemPool.Len() == 0 {
		return nil, nil
	}

	// this is the last slot, so it has to open up progression
	var initialCount int
	if slotPool.Len() == numUsedSlots+1 && !fillUnused {
		initialCount = countFunc(r)
	}

	// try placing an item in the first slot until one fits
	for es := slotPool.Front(); es != nil; es = es.Next() {
		slot := es.Value.(*graph.Node)

		r.Graph.ClearMarks()
		if slot.GetMark(slot, false) != graph.MarkTrue ||
			!canAffordSlot(r, slot) {
			continue
		}

		for ei := itemPool.Front(); ei != nil; ei = ei.Next() {
			item := ei.Value.(*graph.Node)

			if !itemFitsInSlot(item, slot, src) {
				continue
			}

			item.AddParents(slot)

			if canSoftlock(r.Graph) != nil {
				item.RemoveParent(slot)
				continue
			}

			if slotPool.Len() == numUsedSlots+1 && !fillUnused {
				newCount := countFunc(r)
				if newCount <= initialCount {
					item.RemoveParent(slot)
					continue
				}
			}

			return ei, es
		}
	}

	return nil, nil
}

// maps should be looped through based on a sorted set of keys (which can be
// reordered before iteration, as long as it's ordered first); otherwise the
// same random seed can yield different results.
func getSortedKeys(g graph.Graph, src *rand.Rand) []string {
	keys := make([]string, 0, len(g))
	for k := range g {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// checks whether the item fits in the slot due to things like seeds only going
// in trees, certain item slots not accomodating sub IDs. this doesn't check
// for softlocks or the availability of the slot and item.
func itemFitsInSlot(itemNode, slotNode *graph.Node, src *rand.Rand) bool {
	slot := rom.ItemSlots[slotNode.Name]

	// gasha seeds and pieces of heart can be placed in either chests or
	// found/gift slots. beyond that, only unique items can be placed in
	// non-chest slots.
	if itemNode.Name == "gasha seed" || itemNode.Name == "piece of heart" {
		if slotNode.Name == "d0 sword chest" ||
			slotNode.Name == "iron shield gift" ||
			!(rom.IsChest(slot) || rom.IsFound(slot)) {
			return false
		}
	} else if (!rom.IsChest(slot) ||
		slotNode.Name == "d0 sword chest" || slotNode.Name == "rod gift") &&
		!rom.TreasureIsUnique[itemNode.Name] {
		return false
	}

	// dummy shop slots 1 and 2 can only hold their vanilla items.
	if slotNode.Name == "village shop 1" && itemNode.Name != "bombs, 10" {
		return false
	}
	if slotNode.Name == "village shop 2" && itemNode.Name != "shop shield L-1" {
		return false
	}
	if itemNode.Name == "shop shield L-1" && slotNode.Name != "village shop 2" {
		return false
	}

	// give proportionally reduced chances of roughly equivalent items
	// appearing in the d0 sword chest.
	if src != nil {
		if slotNode.Name == "d0 sword chest" {
			switch itemNode.Name {
			case "sword 1", "sword 2":
				if src.Intn(2) != 0 {
					return false
				}
			case "feather 1", "feather 2":
				if src.Intn(2) != 0 {
					return false
				}
			case "winter", "spring", "summer", "autumn":
				if src.Intn(4) != 0 {
					return false
				}
			}
		}
	}

	// these slots won't give you the item if you already have one with that
	// ID, so only use items that have unique IDs and can't be lost.
	switch slotNode.Name {
	case "diver gift", "subrosian market 5", "village shop 3", "star ore spot":
		if !rom.TreasureHasUniqueID(itemNode.Name) ||
			rom.TreasureCanBeLost(itemNode.Name) {
			return false
		}
	}

	// rod of seasons has special graphics something
	if slotNode.Name == "rod gift" && !rom.CanSlotAsRod(itemNode.Name) {
		return false
	}

	// and only seeds can be slotted in seed trees, of course
	switch itemNode.Name {
	case "ember tree seeds", "mystery tree seeds", "scent tree seeds",
		"pegasus tree seeds", "gale tree seeds 1", "gale tree seeds 2":
		switch slotNode.Name {
		case "ember tree", "mystery tree", "scent tree",
			"pegasus tree", "sunken gale tree", "tarm gale tree":
			break
		default:
			return false
		}
	default:
		switch slotNode.Name {
		case "ember tree", "mystery tree", "scent tree",
			"pegasus tree", "sunken gale tree", "tarm gale tree":
			return false
		}
	}

	return true
}

func canAffordSlot(r *Route, slot *graph.Node) bool {
	// if it doesn't cost anything, of course it's affordable
	balance := logic.Rupees[slot.Name]
	if balance >= 0 {
		return true
	}

	// otherwise, count the net rupees available to the player
	balance += r.Costs
	for _, node := range r.Graph {
		value := logic.Rupees[node.Name]
		if value > 0 && node.GetMark(node, false) == graph.MarkTrue {
			balance += value
		}
	}

	return balance >= 0
}
