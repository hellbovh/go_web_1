package homework

import "testing"

func mustNotPanic(t *testing.T, what string, call func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s паникует: %v", what, r)
		}
	}()
	call()
}

func stock() []Item {
	return []Item{{SKU: "A-1", Qty: 10}, {SKU: "B-2", Qty: 3}, {SKU: "A-1", Qty: 5}}
}

func TestReserveChangesCallersSlice(t *testing.T) {
	items := stock()

	if !Reserve(items, "A-1", 4) {
		t.Fatalf("Reserve вернула false, хотя на A-1 хватает остатка")
	}
	if items[0].Qty != 6 {
		t.Errorf("после Reserve items[0].Qty = %d, ожидали 6: "+
			"функция уменьшила копию элемента, а не элемент среза", items[0].Qty)
	}
}

func TestReserveTakesFirstMatchOnly(t *testing.T) {
	items := stock()

	Reserve(items, "A-1", 4)

	if items[2].Qty != 5 {
		t.Errorf("Reserve тронула вторую позицию с тем же артикулом: items[2].Qty = %d, ожидали 5",
			items[2].Qty)
	}
}

func TestReserveRefusesWhenNotEnough(t *testing.T) {
	items := stock()

	if Reserve(items, "B-2", 4) {
		t.Errorf("Reserve вернула true, хотя на B-2 всего 3")
	}
	if items[1].Qty != 3 {
		t.Errorf("отказавшись, Reserve всё равно изменила остаток: items[1].Qty = %d, ожидали 3",
			items[1].Qty)
	}
}

func TestReserveRefusesUnknownSKU(t *testing.T) {
	items := stock()

	if Reserve(items, "нет такого", 1) {
		t.Errorf("Reserve вернула true для несуществующего артикула")
	}
}

func TestSnapshotIsIndependentFromSource(t *testing.T) {
	items := stock()

	copied := Snapshot(items)
	copied[0].Qty = 999

	if items[0].Qty != 10 {
		t.Errorf("правка копии изменила оригинал: items[0].Qty = %d, ожидали 10", items[0].Qty)
	}
}

func TestSnapshotSurvivesChangesInSource(t *testing.T) {
	items := stock()

	copied := Snapshot(items)
	items[1].Qty = 0

	if copied[1].Qty != 3 {
		t.Errorf("правка оригинала изменила копию: copied[1].Qty = %d, ожидали 3", copied[1].Qty)
	}
}

func TestSnapshotKeepsContent(t *testing.T) {
	items := stock()

	copied := Snapshot(items)

	if len(copied) != len(items) {
		t.Fatalf("len(Snapshot) = %d, ожидали %d", len(copied), len(items))
	}
	for i := range items {
		if copied[i] != items[i] {
			t.Errorf("copied[%d] = %+v, ожидали %+v", i, copied[i], items[i])
		}
	}
}

func TestGroupBySKUSumsQuantities(t *testing.T) {
	var total map[string]int

	mustNotPanic(t, "GroupBySKU на непустом складе", func() {
		total = GroupBySKU(stock())
	})

	if total["A-1"] != 15 {
		t.Errorf("total[\"A-1\"] = %d, ожидали 15", total["A-1"])
	}
	if total["B-2"] != 3 {
		t.Errorf("total[\"B-2\"] = %d, ожидали 3", total["B-2"])
	}
}

func TestGroupBySKUOnEmptyInputIsWritable(t *testing.T) {
	var total map[string]int

	mustNotPanic(t, "GroupBySKU на nil-срезе", func() {
		total = GroupBySKU(nil)
	})
	mustNotPanic(t, "запись в результат GroupBySKU", func() {
		total["C-3"] = 1
	})

	if total["C-3"] != 1 {
		t.Errorf("в результат GroupBySKU не удалось записать: total[\"C-3\"] = %d, ожидали 1",
			total["C-3"])
	}
}
