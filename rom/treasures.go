package rom

import (
	"bytes"
	"fmt"
)

// collection modes
// i don't know what the difference between the two find modes are
const (
	CollectGoronGift  = 0x02 // for the L-2 ring box only ??
	CollectUnderwater = 0x08
	CollectFind1      = 0x09
	CollectFind2      = 0x0a
	CollectFall       = 0x29
	CollectChest      = 0x38
	CollectDig        = 0x5a
)

// A Treasure is data associated with a particular item ID and sub ID.
type Treasure struct {
	id, subID byte
	addr      uint16 // bank 15, value of hl at $15:466b

	// in order, starting at addr - 1
	mode, value, text, sprite byte
}

// SubID returns item sub ID of the treasure.
func (t Treasure) SubID() byte {
	return t.subID
}

func (t Treasure) CollectMode() byte {
	return t.mode
}

// RealAddr returns the total offset of the treasure data in the ROM.
func (t Treasure) RealAddr() int {
	return (&Addr{0x15, t.addr}).FullOffset() - 1
}

// Bytes returns a slice of consecutive bytes of treasure data, as they would
// appear in the ROM.
func (t Treasure) Bytes() []byte {
	return []byte{t.mode, t.value, t.text, t.sprite}
}

// Mutate replaces the associated treasure in the given ROM data with this one.
func (t Treasure) Mutate(b []byte) error {
	// fake treasure
	if t.addr == 0 {
		return nil
	}

	addr, data := t.RealAddr(), t.Bytes()
	for i := 0; i < 4; i++ {
		b[addr+i] = data[i]
	}
	return nil
}

// Check verifies that the treasure's data matches the given ROM data.
func (t Treasure) Check(b []byte) error {
	addr, data := t.RealAddr(), t.Bytes()
	if bytes.Compare(b[addr:addr+4], data) != 0 {
		return fmt.Errorf("expected %x at %x; found %x",
			data, addr, b[addr:addr+4])
	}
	return nil
}

// Treasures maps item names to associated treasure data.
var Treasures = map[string]*Treasure{
	"shield L-1":    &Treasure{0x01, 0x00, 0x5701, 0x0a, 0x01, 0x1f, 0x13},
	"bombs":         &Treasure{0x03, 0x00, 0x570d, 0x38, 0x10, 0x4d, 0x05},
	"sword L-1":     &Treasure{0x05, 0x00, 0x571d, 0x38, 0x01, 0x1c, 0x10},
	"sword L-2":     &Treasure{0x05, 0x01, 0x5721, 0x09, 0x02, 0x1d, 0x11},
	"boomerang L-1": &Treasure{0x06, 0x00, 0x5735, 0x0a, 0x01, 0x22, 0x1c},
	"boomerang L-2": &Treasure{0x06, 0x01, 0x5739, 0x38, 0x02, 0x23, 0x1d},
	"rod":           &Treasure{0x07, 0x00, 0x573d, 0x38, 0x07, 0x0a, 0x1e},
	"magnet gloves": &Treasure{0x08, 0x00, 0x558d, 0x38, 0x00, 0x30, 0x18},
	"bombchus":      &Treasure{0x0d, 0x00, 0x5761, 0x0a, 0x10, 0x32, 0x24},
	"strange flute": &Treasure{0x0e, 0x00, 0x55a5, 0x0a, 0x0c, 0x3b, 0x23},
	"slingshot L-1": &Treasure{0x13, 0x00, 0x5769, 0x38, 0x01, 0x2e, 0x21},
	"slingshot L-2": &Treasure{0x13, 0x01, 0x576d, 0x38, 0x02, 0x2f, 0x22},
	"shovel":        &Treasure{0x15, 0x00, 0x55c1, 0x0a, 0x00, 0x25, 0x1b},
	"bracelet":      &Treasure{0x16, 0x00, 0x55c5, 0x38, 0x00, 0x26, 0x19},
	"feather L-1":   &Treasure{0x17, 0x00, 0x5771, 0x38, 0x01, 0x27, 0x16},
	"feather L-2":   &Treasure{0x17, 0x01, 0x5775, 0x38, 0x02, 0x28, 0x17},
	"satchel":       &Treasure{0x19, 0x00, 0x56f9, 0x0a, 0x01, 0x2d, 0x20},
	"fool's ore":    &Treasure{0x1e, 0x00, 0x55e5, 0x00, 0x00, 0xff, 0x1a},
	"flippers":      &Treasure{0x2e, 0x00, 0x5625, 0x0a, 0x00, 0x31, 0x31},

	// rings actually have various entries based on param. this is the first
	// "ring" in the treasure code, but it has a param of 0xff (nothing).
	"ring": &Treasure{0x2d, 0x00, 0x57fd, 0x09, 0xff, 0x54, 0x0e},

	// ignore these ones for the purpose of rom validation, since they're going
	// to be custom from the start.
	//
	// XXX disable these, at least temporarily, since they can cause softlocks
	//     if they appear in non-chest locations *and* normal chests. it's
	//     probably possible to use the first four unused(?) ring slots for
	//     these instead.
	/*
		"find expert's ring": &Treasure{0x2d, 0x04, 0x580d, 0x38, 0x0b, 0x54, 0x0e},
		"find toss ring":     &Treasure{0x2d, 0x05, 0x5811, 0x38, 0x12, 0x54, 0x0e},
		"find energy ring":   &Treasure{0x2d, 0x06, 0x5815, 0x38, 0x31, 0x54, 0x0e},
		"find fist ring":     &Treasure{0x2d, 0x07, 0x5819, 0x38, 0x3d, 0x54, 0x0e},
	*/

	// the keys have several params that indicate which mode to use (fall,
	// appear, exist, chest?). the compass and map have the same, but without
	// a fall mode.
	"small key": &Treasure{0x30, 0x00, 0x584d, 0x38, 0x01, 0x1a, 0x42},
	"boss key":  &Treasure{0x31, 0x00, 0x585d, 0x38, 0x00, 0x1b, 0x43},
	"compass":   &Treasure{0x32, 0x00, 0x5869, 0x68, 0x00, 0x19, 0x41},
	"map":       &Treasure{0x33, 0x00, 0x5875, 0x68, 0x00, 0x18, 0x40},

	"gnarled key":     &Treasure{0x42, 0x00, 0x58a9, 0x29, 0x00, 0x42, 0x44},
	"ricky's gloves":  &Treasure{0x48, 0x00, 0x568d, 0x09, 0x01, 0x67, 0x55},
	"floodgate key":   &Treasure{0x43, 0x00, 0x5679, 0x09, 0x00, 0x43, 0x45},
	"star ore":        &Treasure{0x45, 0x00, 0x5681, 0x5a, 0x00, 0x40, 0x57},
	"square jewel":    &Treasure{0x4e, 0x00, 0x56a5, 0x38, 0x00, 0x48, 0x38},
	"master's plaque": &Treasure{0x54, 0x00, 0x56bd, 0x38, 0x00, 0x70, 0x26},
	"spring banana":   &Treasure{0x47, 0x00, 0x5689, 0x0a, 0x00, 0x66, 0x54},
	"dragon key":      &Treasure{0x44, 0x00, 0x567d, 0x09, 0x00, 0x44, 0x46},
	"pyramid jewel":   &Treasure{0x4d, 0x00, 0x58bd, 0x08, 0x00, 0x4a, 0x37},
	"x-shaped jewel":  &Treasure{0x4f, 0x00, 0x56a9, 0x38, 0x00, 0x49, 0x39},
	"round jewel":     &Treasure{0x4c, 0x00, 0x569d, 0x0a, 0x00, 0x47, 0x36},
	"rusty bell":      &Treasure{0x4a, 0x00, 0x58b1, 0x0a, 0x00, 0x55, 0x5b},
	"ring box L-2":    &Treasure{0x2c, 0x02, 0x57f1, 0x02, 0x03, 0x34, 0x35},

	// this one doesn't seem to work as it should for its collect mode
	//"piece of heart (find)":  &Treasure{0x2b, 0x01, 0x57d1, 0x0a, 0x01, 0x17, 0x3a},

	"piece of heart (chest)": &Treasure{0x2b, 0x01, 0x57d5, 0x38, 0x01, 0x17, 0x3a},

	// these seeds are "fake" treasures. real treasures corresponding to each
	// type of seed exist, but those can't be used for changing which tree
	// yields which seeds.
	"ember tree seeds":   &Treasure{id: 0x00},
	"mystery tree seeds": &Treasure{id: 0x01},
	"scent tree seeds":   &Treasure{id: 0x02},
	"pegasus tree seeds": &Treasure{id: 0x03},
	"gale tree seeds 1":  &Treasure{id: 0x04},
	"gale tree seeds 2":  &Treasure{id: 0x05},
}

var seedIndexByTreeID = []byte{0, 4, 1, 2, 3, 3}

// reverse lookup the treasure name; returns empty string if not found. this
// ignores fake seed treasures.
func treasureNameFromIDs(id, subID byte) string {
	for k, v := range Treasures {
		if v.addr != 0 && v.id == id && v.subID == subID {
			return k
		}
	}
	return ""
}
