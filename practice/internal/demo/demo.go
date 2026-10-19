// Package demo — наблюдения, на которых держится теория вебинара.
//
// Здесь нет ни одного заранее написанного числа и ни одной заранее написанной
// строки вывода: всё, что печатает пакет, он сначала делает, а потом показывает
// результат. Слайды вебинара берут эти же значения из отчёта — поэтому строка
// на слайде и строка в терминале студента совпадают не по договорённости,
// а потому что это один и тот же прогон.
//
// Наблюдения выбраны так, чтобы результат не зависел от версии Go и от машины:
// адреса не печатаются, вместимость после роста среза не печатается (алгоритм
// роста — деталь реализации рантайма), поведение append за пределами cap
// проверяется вопросом «остался ли общий массив», а не числом.
package demo

import (
	"fmt"
	"unicode/utf8"
)

// Section — один блок вывода: заголовок и уже посчитанные строки.
type Section struct {
	Key   string   `json:"key"`
	Title string   `json:"title"`
	Lines []string `json:"lines"`
}

// Result — всё, что дал прогон: текст для экрана и измеренные значения.
type Result struct {
	Sections []Section      `json:"sections"`
	Facts    map[string]any `json:"facts"`
}

// Counter — та же структура, что в задаче 1: два поля подряд.
type Counter struct {
	Name string
	Hits int
}

// Пример строки для блока про байты и руны. Взят кириллический текст с длинным
// тире: в нём каждая из трёх категорий символов занимает разное число байт.
const sample = "Go — просто"

// Run выполняет все наблюдения и возвращает их результат.
func Run() Result {
	facts := map[string]any{}
	sections := []Section{
		copySection(facts),
		pointerSection(facts),
		rangeSection(facts),
		arrayVsSliceSection(facts),
		sliceHeaderSection(facts),
		regrowSection(facts),
		stringSection(facts),
		zeroSection(facts),
	}
	return Result{Sections: sections, Facts: facts}
}

// Lines разворачивает результат в плоский текст для терминала.
func (r Result) Lines() []string {
	out := make([]string, 0, len(r.Sections)*6)
	for i, section := range r.Sections {
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, fmt.Sprintf("[%d] %s", i+1, section.Title))
		out = append(out, section.Lines...)
	}
	return out
}

// Section возвращает блок по ключу; второе значение — был ли такой блок.
func (r Result) Section(key string) (Section, bool) {
	for _, section := range r.Sections {
		if section.Key == key {
			return section, true
		}
	}
	return Section{}, false
}

func copySection(facts map[string]any) Section {
	c := Counter{Name: "api", Hits: 0}
	d := c
	d.Hits = 5

	facts["copy_original_hits"] = c.Hits
	facts["copy_copy_hits"] = d.Hits
	return Section{
		Key:   "copy",
		Title: "присваивание структуры копирует все её поля",
		Lines: []string{
			`c := Counter{Name: "api"}`,
			"d := c;  d.Hits = 5",
			fmt.Sprintf("c = %+v", c),
			fmt.Sprintf("d = %+v", d),
		},
	}
}

func pointerSection(facts map[string]any) Section {
	c := Counter{Name: "api", Hits: 0}
	p := &c
	p.Hits = 5

	facts["pointer_original_hits"] = c.Hits
	return Section{
		Key:   "pointer",
		Title: "в указателе лежит адрес, и запись по нему меняет оригинал",
		Lines: []string{
			"p := &c;  p.Hits = 5",
			fmt.Sprintf("c = %+v", c),
		},
	}
}

func rangeSection(facts map[string]any) Section {
	day := []Counter{{Name: "api"}, {Name: "web"}}
	for _, item := range day {
		item.Hits++
	}
	afterValueForm := day[0].Hits

	for i := range day {
		day[i].Hits++
	}
	afterIndexForm := day[0].Hits

	facts["range_value_form_hits"] = afterValueForm
	facts["range_index_form_hits"] = afterIndexForm
	return Section{
		Key:   "range",
		Title: "range со вторым значением отдаёт копию элемента",
		Lines: []string{
			"for _, item := range day { item.Hits++ }",
			fmt.Sprintf("day[0].Hits = %d", afterValueForm),
			"for i := range day       { day[i].Hits++ }",
			fmt.Sprintf("day[0].Hits = %d", afterIndexForm),
		},
	}
}

func arrayVsSliceSection(facts map[string]any) Section {
	arr := [3]int{1, 2, 3}
	arrCopy := arr
	arrCopy[0] = 9

	sl := []int{1, 2, 3}
	slAlias := sl
	slAlias[0] = 9

	facts["array_original_first"] = arr[0]
	facts["slice_original_first"] = sl[0]
	return Section{
		Key:   "array_vs_slice",
		Title: "массив копируется целиком, срез — нет",
		Lines: []string{
			"arr := [3]int{1, 2, 3};  cp := arr;    cp[0] = 9",
			fmt.Sprintf("arr = %v", arr),
			"sl  := []int{1, 2, 3};   alias := sl;  alias[0] = 9",
			fmt.Sprintf("sl  = %v", sl),
		},
	}
}

func sliceHeaderSection(facts map[string]any) Section {
	base := []int{1, 2, 3, 4}
	head := base[:2]
	headLen, headCap := len(head), cap(head)
	shared := &head[0] == &base[0]

	head = append(head, 99)

	facts["head_len"] = headLen
	facts["head_cap"] = headCap
	facts["head_shares_base"] = shared
	facts["base_after_append"] = fmt.Sprint(base)
	return Section{
		Key:   "header",
		Title: "срез хранит указатель на массив, длину и вместимость",
		Lines: []string{
			"base := []int{1, 2, 3, 4};  head := base[:2]",
			fmt.Sprintf("len(head) = %d   cap(head) = %d   общий массив: %v",
				headLen, headCap, shared),
			"head = append(head, 99)",
			fmt.Sprintf("base = %v", base),
		},
	}
}

func regrowSection(facts map[string]any) Section {
	base := []int{1, 2, 3, 4}
	full := base[:len(base)]
	const extra = 5
	grown := append(full, extra)
	shared := &grown[0] == &base[0]

	facts["regrow_appended"] = extra
	facts["regrow_shares_base"] = shared
	facts["regrow_len"] = len(grown)
	facts["regrow_base"] = fmt.Sprint(base)
	return Section{
		Key:   "regrow",
		Title: "append за пределами вместимости заводит новый массив",
		Lines: []string{
			fmt.Sprintf("full := base[:%d];  grown := append(full, %d)", len(base), extra),
			fmt.Sprintf("len(grown) = %d   общий массив с base: %v", len(grown), shared),
			fmt.Sprintf("base = %v   grown = %v", base, grown),
		},
	}
}

func stringSection(facts map[string]any) Section {
	bytes := len(sample)
	runes := utf8.RuneCountInString(sample)
	first := sample[0]

	facts["string_bytes"] = bytes
	facts["string_runes"] = runes
	facts["string_first_byte"] = int(first)
	facts["string_sample"] = sample
	return Section{
		Key:   "string",
		Title: "в строке лежат байты, а не символы",
		Lines: []string{
			fmt.Sprintf("s := %q", sample),
			fmt.Sprintf("len(s) = %d   рун = %d   s[0] = %d", bytes, runes, first),
		},
	}
}

func zeroSection(facts map[string]any) Section {
	var xs []int
	appended := append(xs, 1)

	var m map[string]int
	read := m["нет такого"]

	panicText := writePanic(m)

	facts["nil_slice_is_nil"] = xs == nil
	facts["nil_slice_len"] = len(xs)
	facts["nil_slice_after_append"] = fmt.Sprint(appended)
	facts["nil_map_read"] = read
	facts["nil_map_write_panic"] = panicText
	return Section{
		Key:   "zero",
		Title: "нулевое значение уже лежит в переменной",
		Lines: []string{
			fmt.Sprintf("var xs []int          xs == nil: %v   len(xs) = %d", xs == nil, len(xs)),
			fmt.Sprintf("append(xs, 1)         %v", appended),
			fmt.Sprintf(`var m map[string]int  m["нет такого"] = %d`, read),
			fmt.Sprintf(`m["есть"] = 1         panic: %s`, panicText),
		},
	}
}

// writePanic выполняет запись в nil-карту и возвращает текст паники.
// Это единственный способ показать сообщение рантайма, не уронив программу.
func writePanic(m map[string]int) (text string) {
	defer func() {
		if r := recover(); r != nil {
			text = fmt.Sprint(r)
		}
	}()
	m["есть"] = 1
	return "паники не было"
}
