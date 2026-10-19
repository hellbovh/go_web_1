package task1values

import "testing"

// Тест проверяет ровно одно: увидел ли изменения тот, кто вызвал функцию.
// Ни одна проверка не заглядывает внутрь реализации — только в свои собственные
// переменные до и после вызова.

func TestHitChangesCallersCounter(t *testing.T) {
	c := Counter{Name: "api", Hits: 0}

	Hit(&c)

	if c.Hits != 1 {
		t.Errorf("после Hit(&c) у вызывающей стороны c.Hits = %d, ожидали 1: "+
			"функция изменила копию, а не сам счётчик", c.Hits)
	}
}

func TestHitIsRepeatable(t *testing.T) {
	c := Counter{Name: "api"}

	Hit(&c)
	Hit(&c)
	Hit(&c)

	if c.Hits != 3 {
		t.Errorf("после трёх вызовов Hit(&c) c.Hits = %d, ожидали 3", c.Hits)
	}
}

func TestHitDoesNotTouchName(t *testing.T) {
	c := Counter{Name: "api", Hits: 7}

	Hit(&c)

	if c.Name != "api" {
		t.Errorf("Hit изменил Name: получили %q, ожидали \"api\"", c.Name)
	}
	if c.Hits != 8 {
		t.Errorf("Hit(&c) при Hits=7 дал %d, ожидали 8", c.Hits)
	}
}

func TestHitAllChangesEveryElement(t *testing.T) {
	day := []Counter{{Name: "api"}, {Name: "web"}, {Name: "admin"}}

	HitAll(day)

	for i := range day {
		if day[i].Hits != 1 {
			t.Errorf("после HitAll у day[%d] (%s) Hits = %d, ожидали 1: "+
				"цикл увеличил копию элемента, а не элемент среза",
				i, day[i].Name, day[i].Hits)
		}
	}
}

func TestHitAllKeepsOrderAndNames(t *testing.T) {
	day := []Counter{{Name: "api", Hits: 4}, {Name: "web", Hits: 9}, {Name: "admin", Hits: 1}}

	HitAll(day)

	want := []Counter{{Name: "api", Hits: 5}, {Name: "web", Hits: 10}, {Name: "admin", Hits: 2}}
	for i := range want {
		if day[i] != want[i] {
			t.Errorf("day[%d] = %+v, ожидали %+v", i, day[i], want[i])
		}
	}
	if got := Busiest(day); got != "web" {
		t.Errorf("Busiest после HitAll = %q, ожидали \"web\"", got)
	}
}

// Этот тест проходит и на стартовом коде. Он здесь не для проверки задачи,
// а как опора для разбора: массив — это значение, и присваивание копирует его
// целиком, сколько бы элементов в нём ни было.
func TestArrayIsCopiedOnAssignment(t *testing.T) {
	original := [3]Counter{{Name: "api"}, {Name: "web"}, {Name: "admin"}}

	copied := original
	copied[0].Hits = 100

	if original[0].Hits != 0 {
		t.Errorf("присваивание массива не скопировало его: original[0].Hits = %d, ожидали 0",
			original[0].Hits)
	}
	if copied[0].Hits != 100 {
		t.Errorf("копия не изменилась: copied[0].Hits = %d, ожидали 100", copied[0].Hits)
	}
}
