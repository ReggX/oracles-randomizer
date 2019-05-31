package randomizer

import (
	"fmt"
)

// an instance of ROM data that can be changed by the randomizer.
type mutable interface {
	mutate([]byte) error // change ROM bytes
	check([]byte) error  // verify that the mutable matches the ROM
}

// a length of mutable bytes starting at a given address.
type mutableRange struct {
	addr     address
	old, new []byte
}

// implements `mutate()` from the `mutable` interface.
func (mut *mutableRange) mutate(b []byte) error {
	offset := mut.addr.fullOffset()
	for i, value := range mut.new {
		b[offset+i] = value
	}
	return nil
}

// implements `check()` from the `mutable` interface.
func (mut *mutableRange) check(b []byte) error {
	offset := mut.addr.fullOffset()
	for i, value := range mut.old {
		if b[offset+i] != value {
			return fmt.Errorf("expected %x at %x; found %x",
				mut.old[i], offset+i, b[offset+i])
		}
	}
	return nil
}

// sets music on or off in the modified ROM. By default, it is off.
func (rom *romState) setMusic(music bool) {
	mut := rom.codeMutables["filterMusic"]
	mut.new[3] = byte(ternary(music, 0x18, 0x30).(int)) // jr / jr nc
}

// sets treewarp on or off in the modified ROM. By default, it is on.
func (rom *romState) setTreewarp(treewarp bool) {
	mut := rom.codeMutables["treeWarp"]
	mut.new[5] = byte(ternary(treewarp, 0x28, 0x18).(int)) // jr z / jr
}

// sets the flute type and natzu region type based on a companion number 1 to
// 3.
func (rom *romState) setAnimal(companion int) {
	rom.codeMutables["romAnimalRegion"].new =
		[]byte{byte(companion + 0x0a)}
	rom.codeMutables["flutePalette"].new =
		[]byte{byte(0x10*(4-companion) + 3)}
}

// key = area name (as in asm/vars.yaml), id = season index (spring -> winter).
func (rom *romState) setSeason(key string, id byte) {
	rom.codeMutables[key].new[0] = id
}

// get a collated map of all mutables, *except* for treasures which do not
// appear in the seed. this allows things like the three seasons flutes having
// different data but the same address.
func (rom *romState) getAllMutables() map[string]mutable {
	allMutables := make(map[string]mutable)
	for k, v := range rom.itemSlots {
		if v.treasure == nil {
			panic(fmt.Sprintf("treasure for %s is nil", k))
		}
		if v.treasure.addr.offset != 0 {
			// treasures are allowed to have duplicates
			tName, _ := reverseLookup(rom.treasures, v.treasure)
			allMutables[tName.(string)] = v.treasure
		}
		addMutOrPanic(allMutables, k, v)
	}
	for k, v := range rom.codeMutables {
		addMutOrPanic(allMutables, k, v)
	}
	return allMutables
}

// if the mutable does not exist in the map, add it. if it already exists,
// panic.
func addMutOrPanic(m map[string]mutable, k string, v mutable) {
	if _, ok := m[k]; ok {
		panic("duplicate mutable key: " + k)
	}
	m[k] = v
}

// returns the name of a mutable that covers the given address, or an empty
// string if none is found.
func (rom *romState) findAddr(bank byte, addr uint16) string {
	muts := rom.getAllMutables()
	offset := (&address{bank, addr}).fullOffset()

	for name, mut := range muts {
		switch mut := mut.(type) {
		case *mutableRange:
			if offset >= mut.addr.fullOffset() &&
				offset < mut.addr.fullOffset()+len(mut.new) {
				return name
			}
		case *itemSlot:
			for _, addrs := range [][]address{mut.idAddrs, mut.subidAddrs} {
				for _, addr := range addrs {
					if offset == addr.fullOffset() {
						return name
					}
				}
			}
		case *treasure:
			if offset >= mut.addr.fullOffset() &&
				offset < mut.addr.fullOffset()+4 {
				return name
			}
		default:
			panic("unknown type for mutable: " + name)
		}
	}

	return ""
}