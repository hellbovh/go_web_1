package task2slices

import "testing"

// Проверки построены так, чтобы общую память было видно без отладчика: берём
// два среза, меняем один и смотрим на другой.

func TestHeadReturnsRequestedPrefix(t *testing.T) {
	src := []int{10, 20, 30, 40}

	got := Head(src, 2)

	if len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Fatalf("Head(src, 2) = %v, ожидали [10 20]", got)
	}
}

func TestHeadDoesNotShareArrayWithSource(t *testing.T) {
	src := []int{10, 20, 30, 40}

	got := Head(src, 2)
	got[0] = 99

	if src[0] != 10 {
		t.Errorf("после записи в результат Head исходный src[0] = %d, ожидали 10: "+
			"результат смотрит в тот же массив, что и src", src[0])
	}
}

func TestHeadResultSurvivesChangesInSource(t *testing.T) {
	src := []int{10, 20, 30, 40}

	got := Head(src, 2)
	src[1] = 77

	if got[1] != 20 {
		t.Errorf("после записи в src результат Head изменился: got[1] = %d, ожидали 20", got[1])
	}
}

func TestHeadOfZeroLength(t *testing.T) {
	src := []int{10, 20}

	got := Head(src, 0)

	if len(got) != 0 {
		t.Errorf("Head(src, 0) вернул %v, ожидали пустой срез", got)
	}
}

// Ключевой тест задачи. base — массив на пять элементов, из которых заняты три;
// tags смотрит только на первые два. Запись append в свободную вместимость
// попадает ровно в base[2] — туда, где уже лежит чужое значение.
func TestAddTagDoesNotOverwriteTail(t *testing.T) {
	base := make([]string, 3, 5)
	copy(base, []string{"prod", "api", "important"})
	tags := base[:2]

	AddTag(tags, "new")

	if base[2] != "important" {
		t.Errorf("AddTag затёр base[2]: получили %q, ожидали \"important\": "+
			"append записал новый элемент в общий массив", base[2])
	}
}

func TestAddTagReturnsExtendedSlice(t *testing.T) {
	tags := []string{"prod", "api"}

	got := AddTag(tags, "new")

	want := []string{"prod", "api", "new"}
	if len(got) != len(want) {
		t.Fatalf("AddTag вернул %v, ожидали %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AddTag вернул %v, ожидали %v", got, want)
		}
	}
}

func TestAddTagLeavesSourceLengthAlone(t *testing.T) {
	tags := []string{"prod", "api"}

	AddTag(tags, "new")

	if len(tags) != 2 {
		t.Errorf("после AddTag len(tags) = %d, ожидали 2", len(tags))
	}
}

func TestAddTagWorksOnNilSlice(t *testing.T) {
	var tags []string

	got := AddTag(tags, "first")

	if len(got) != 1 || got[0] != "first" {
		t.Errorf("AddTag(nil, \"first\") = %v, ожидали [first]", got)
	}
}

func TestSharesArraySeesCommonArray(t *testing.T) {
	base := []int{1, 2, 3, 4}
	head := base[:2]

	if !SharesArray(base, head) {
		t.Errorf("SharesArray(base, base[:2]) = false, ожидали true: " +
			"оба среза начинаются с одного элемента массива")
	}
}

func TestSharesArraySeesIndependentSlices(t *testing.T) {
	base := []int{1, 2, 3, 4}
	other := []int{1, 2, 3, 4}

	if SharesArray(base, other) {
		t.Errorf("SharesArray для двух разных срезов с одинаковым содержимым = true, " +
			"ожидали false: одинаковые значения не означают общий массив")
	}
}

func TestSharesArrayOnEmptySlices(t *testing.T) {
	var empty []int

	if SharesArray(empty, empty) {
		t.Errorf("SharesArray для пустых срезов = true, ожидали false: " +
			"у пустого среза нет элемента, адрес которого можно сравнить")
	}
}

// Тест проходит и на стартовом коде. Он фиксирует то, ради чего задача вообще
// существует: append за пределами вместимости выделяет новый массив, и связь
// между срезами при этом рвётся.
func TestAppendBeyondCapacityBreaksSharing(t *testing.T) {
	base := []int{1, 2, 3, 4}
	full := base[:len(base)]

	grown := append(full, 5)

	if len(grown) != 5 {
		t.Fatalf("len(grown) = %d, ожидали 5", len(grown))
	}
	if &grown[0] == &base[0] {
		t.Errorf("после append за пределы cap срез остался в старом массиве: " +
			"этого не может быть, места для пятого элемента там нет")
	}
	if base[0] != 1 {
		t.Errorf("append изменил base[0] = %d, ожидали 1", base[0])
	}
}
