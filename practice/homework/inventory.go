// Package homework — домашнее задание вебинара «Что лежит в переменной».
//
// Это не новый проект, а продолжение того же: три функции, по одной на каждую
// задачу эфира. Ошибки в стартовом коде здесь ровно те же самые, только
// спрятаны в задачу поинтереснее — учёт остатков на складе.
//
// Запуск: go test ./homework
package homework

// Item — позиция на складе: артикул и количество.
type Item struct {
	SKU string
	Qty int
}

// Reserve уменьшает количество у первой позиции с артикулом sku на qty
// и возвращает true.
//
// Если такой позиции нет или на ней меньше qty, ничего не меняется и
// возвращается false. Изменение обязано быть видно вызывающей стороне.
func Reserve(items []Item, sku string, qty int) bool {
	// TODO(домашнее задание): функция повторяет одну из ошибок эфира. Какую —
	// подскажет сообщение упавшего теста: go test ./homework
	// for _, item := range items {
	//	if item.SKU == sku && item.Qty >= qty {
	//		item.Qty -= qty
	//		return true
	//	}
	// }
	// return false
	for i := range items {
		if items[i].SKU == sku && items[i].Qty >= qty {
			items[i].Qty -= qty
			return true
		}
	}
	return false
}

// Snapshot возвращает копию списка позиций, живущую своей жизнью.
//
// Правки копии не должны быть видны в оригинале, а правки оригинала — в копии.
func Snapshot(items []Item) []Item {
	// TODO(домашнее задание): функция повторяет одну из ошибок эфира. Какую —
	// подскажет сообщение упавшего теста: go test ./homework
	// return items
	res := make([]Item, len(items))
	copy(res, items)
	return res
}

// GroupBySKU возвращает суммарное количество по каждому артикулу.
//
// Результат обязан быть картой, в которую вызывающая сторона может писать
// сразу, — даже если items пуст или равен nil.
func GroupBySKU(items []Item) map[string]int {
	// TODO(домашнее задание): функция повторяет одну из ошибок эфира. Какую —
	// подскажет сообщение упавшего теста: go test ./homework
	// var total map[string]int
	//for _, item := range items {
	//	total[item.SKU] += item.Qty
	// }
	// return total
	total := make(map[string]int)
	for _, item := range items {
		total[item.SKU] += item.Qty
	}
	return total	
}
