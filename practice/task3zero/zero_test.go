package task3zero

import "testing"

// mustNotPanic выполняет call и превращает панику в понятный провал теста.
// Без него первая же паника уронила бы весь тестовый бинарь пакета, и остальные
// проверки не успели бы ничего рассказать.
func mustNotPanic(t *testing.T, what string, call func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s паникует: %v", what, r)
		}
	}()
	call()
}

func TestNewProfileKeepsName(t *testing.T) {
	p := NewProfile("ada")

	if p.Name != "ada" {
		t.Errorf("NewProfile(\"ada\").Name = %q, ожидали \"ada\"", p.Name)
	}
}

func TestNewProfileIsReadyForWrites(t *testing.T) {
	p := NewProfile("ada")

	mustNotPanic(t, "запись в Meta профиля из NewProfile", func() {
		p.Meta["role"] = "admin"
	})

	if p.Meta["role"] != "admin" {
		t.Errorf("после записи Meta[\"role\"] = %q, ожидали \"admin\"", p.Meta["role"])
	}
}

func TestNewProfileLeavesTagsNil(t *testing.T) {
	p := NewProfile("ada")

	if p.Tags != nil {
		t.Errorf("NewProfile завёл пустой срез Tags = %v: nil-срез уже готов к append, "+
			"создавать его заранее незачем", p.Tags)
	}
	appended := append(p.Tags, "new")
	if len(appended) != 1 || appended[0] != "new" {
		t.Errorf("append к nil-срезу Tags дал %v, ожидали [new]", appended)
	}
}

// Ключевой тест задачи: профиль объявлен через var, никакого конструктора не было,
// Meta в нём nil. Это ровно тот случай, в котором ломается неаккуратная SetMeta.
func TestSetMetaWorksOnZeroProfile(t *testing.T) {
	var p Profile

	mustNotPanic(t, "SetMeta на профиле с nil-картой", func() {
		SetMeta(&p, "role", "admin")
	})

	if p.Meta["role"] != "admin" {
		t.Errorf("после SetMeta значение Meta[\"role\"] = %q, ожидали \"admin\"", p.Meta["role"])
	}
}

func TestSetMetaKeepsExistingKeys(t *testing.T) {
	p := Profile{Name: "ada", Meta: map[string]string{"role": "admin"}}

	SetMeta(&p, "team", "platform")

	if p.Meta["role"] != "admin" {
		t.Errorf("SetMeta потеряла существующий ключ: Meta[\"role\"] = %q, ожидали \"admin\"",
			p.Meta["role"])
	}
	if p.Meta["team"] != "platform" {
		t.Errorf("Meta[\"team\"] = %q, ожидали \"platform\"", p.Meta["team"])
	}
}

func TestSetMetaOverwritesValue(t *testing.T) {
	var p Profile

	mustNotPanic(t, "повторная SetMeta на профиле с nil-картой", func() {
		SetMeta(&p, "role", "admin")
		SetMeta(&p, "role", "reader")
	})

	if p.Meta["role"] != "reader" {
		t.Errorf("после повторной SetMeta Meta[\"role\"] = %q, ожидали \"reader\"", p.Meta["role"])
	}
}

func TestLookupFindsExistingKey(t *testing.T) {
	p := Profile{Meta: map[string]string{"role": "admin"}}

	value, ok := Lookup(p, "role")

	if value != "admin" || !ok {
		t.Errorf("Lookup(p, \"role\") = (%q, %v), ожидали (\"admin\", true)", value, ok)
	}
}

// Значение ключа — пустая строка. Это не то же самое, что «ключа нет»:
// сама пустая строка про наличие ключа ничего не сообщает.
func TestLookupDistinguishesEmptyValueFromMissingKey(t *testing.T) {
	p := Profile{Meta: map[string]string{"note": ""}}

	value, ok := Lookup(p, "note")

	if value != "" {
		t.Errorf("Lookup вернул значение %q, ожидали пустую строку", value)
	}
	if !ok {
		t.Errorf("Lookup сообщил, что ключа \"note\" нет, а он есть и хранит пустую строку")
	}
}

func TestLookupReportsMissingKey(t *testing.T) {
	p := Profile{Meta: map[string]string{"role": "admin"}}

	value, ok := Lookup(p, "team")

	if ok {
		t.Errorf("Lookup сообщил, что ключ \"team\" есть, а его нет")
	}
	if value != "" {
		t.Errorf("для отсутствующего ключа Lookup вернул %q, ожидали пустую строку", value)
	}
}

func TestLookupOnNilMapDoesNotPanic(t *testing.T) {
	var p Profile

	mustNotPanic(t, "чтение из nil-карты", func() {
		value, ok := Lookup(p, "role")
		if ok {
			t.Errorf("Lookup на nil-карте сообщил о наличии ключа, значение %q", value)
		}
	})
}
