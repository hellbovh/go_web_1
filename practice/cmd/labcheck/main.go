// Command labcheck прогоняет тесты всех задач вебинара и показывает, что уже
// сходится, а что нет.
//
//	go run ./cmd/labcheck                       # ваши задачи и эталон, если доступен
//	go run ./cmd/labcheck -reference=""         # не трогать эталон
//	go run ./cmd/labcheck -report ../build/student_report.json
//
// Программа не заменяет `go test ./...` и ничего не знает про правильные ответы:
// она запускает тот же самый `go test`, читает его машинный вывод и складывает
// результат в один отчёт. Второй прогон — по эталонному решению из ../reference:
// эталонные файлы подменяют ваши во временной копии проекта, ваши файлы при этом
// не трогаются.
//
// Отчёт в формате JSON использует сборка вебинара: числа на слайдах приезжают
// из него, а не пишутся руками.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gobasics/variables/internal/demo"
)

// modulePath — путь модуля из go.mod. Нужен, чтобы сокращать имена пакетов
// в отчёте до директорий: gobasics/variables/task1values -> task1values.
const modulePath = "gobasics/variables"

// TestResult — исход одного теста.
type TestResult struct {
	Package  string   `json:"package"`
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Messages []string `json:"messages,omitempty"`
}

// Suite — исход одного прогона тестов целиком.
type Suite struct {
	Tests       []TestResult `json:"tests"`
	Passed      int          `json:"passed"`
	Failed      int          `json:"failed"`
	Total       int          `json:"total"`
	FailedNames []string     `json:"failed_names"`
}

// GoInfo — тулчейн, на котором получен отчёт.
type GoInfo struct {
	Version string `json:"version"`
	GOOS    string `json:"goos"`
	GOARCH  string `json:"goarch"`
}

// Report — то, что программа печатает в JSON.
type Report struct {
	Go        GoInfo              `json:"go"`
	Packages  []string            `json:"packages"`
	Starter   Suite               `json:"starter"`
	Reference *Suite              `json:"reference,omitempty"`
	Demo      []demo.Section      `json:"demo"`
	DemoLines map[string][]string `json:"demo_lines"`
	Facts     map[string]any      `json:"facts"`
}

func main() {
	reference := flag.String("reference", "../reference",
		"директория с эталонным решением; пустая строка — не проверять эталон")
	reportPath := flag.String("report", "", "куда записать отчёт в формате JSON")
	flag.Parse()

	if err := run(*reference, *reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "labcheck: %v\n", err)
		os.Exit(1)
	}
}

func run(referenceDir, reportPath string) error {
	packages, err := taskPackages(".")
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return fmt.Errorf("в проекте нет ни одной директории задачи (task*)")
	}

	result := demo.Run()
	report := Report{
		Go:        GoInfo{Version: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH},
		Packages:  packages,
		Demo:      result.Sections,
		DemoLines: demoLines(result.Sections),
		Facts:     result.Facts,
	}

	starter, err := runTests(".", packages)
	if err != nil {
		return fmt.Errorf("прогон тестов задач: %w", err)
	}
	report.Starter = starter
	printSuite("ваше решение", starter)

	if referenceDir != "" {
		if _, err := os.Stat(referenceDir); os.IsNotExist(err) {
			// В студенческом архиве эталон намеренно не распространяется.
			fmt.Printf("\nэталон %q отсутствует: проверяем только студенческие задачи\n", referenceDir)
		} else if err != nil {
			return fmt.Errorf("проверка каталога эталона %q: %w", referenceDir, err)
		} else {
			suite, err := runReference(referenceDir, packages)
			if err != nil {
				return fmt.Errorf("прогон эталонного решения: %w", err)
			}
			report.Reference = suite
			printSuite("эталон из "+referenceDir, *suite)
		}
	}

	if reportPath != "" {
		if err := writeReport(reportPath, report); err != nil {
			return err
		}
		fmt.Printf("\nотчёт записан: %s\n", reportPath)
	}
	return nil
}

// demoLines раскладывает блоки демонстрации в плоскую карту «ключ -> строки».
// Она нужна проверке материалов: сравнивать вывод по ключу удобнее, чем искать
// нужный блок в списке.
func demoLines(sections []demo.Section) map[string][]string {
	lines := make(map[string][]string, len(sections))
	for _, section := range sections {
		lines[section.Key] = section.Lines
	}
	return lines
}

// taskPackages находит в корне модуля директории, у которых есть тесты.
//
// Признак — файл _test.go внутри, а не имя директории: так в проверку сразу
// попадает и домашнее задание, и любая задача, которая появится позже, без
// правки этого файла. Служебные пакеты (cmd, internal) тестов не содержат
// и в список не попадают.
func taskPackages(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var packages []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		tested, err := hasTests(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		if tested {
			packages = append(packages, "./"+entry.Name())
		}
	}
	sort.Strings(packages)
	return packages, nil
}

func hasTests(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			return true, nil
		}
	}
	return false, nil
}

// runTests запускает `go test -json` в dir и разбирает машинный вывод.
//
// Ненулевой код возврата здесь — это нормальный исход: до того, как задачи
// решены, тесты обязаны падать. Ошибкой считается только то, что вывод не
// удалось разобрать или тесты не запустились вовсе.
func runTests(dir string, packages []string) (Suite, error) {
	args := append([]string{"test", "-json", "-count=1"}, packages...)
	command := exec.Command("go", args...)
	command.Dir = dir
	command.Stderr = os.Stderr
	output, err := command.Output()
	if len(output) == 0 {
		return Suite{}, fmt.Errorf("go test не дал вывода: %w", err)
	}
	return parseTestJSON(output)
}

// event — одна запись машинного вывода `go test -json`.
type event struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

func parseTestJSON(raw []byte) (Suite, error) {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	type key struct{ pkg, test string }
	statuses := map[key]string{}
	messages := map[key][]string{}
	var order []key

	for decoder.More() {
		var e event
		if err := decoder.Decode(&e); err != nil {
			return Suite{}, fmt.Errorf("не удалось разобрать вывод go test: %w", err)
		}
		if e.Test == "" {
			continue // события уровня пакета: сборка, итог, время прогона
		}
		id := key{shortPackage(e.Package), e.Test}
		switch e.Action {
		case "run":
			if _, seen := statuses[id]; !seen {
				statuses[id] = "run"
				order = append(order, id)
			}
		case "pass", "fail", "skip":
			statuses[id] = e.Action
		case "output":
			if line, ok := assertionLine(e.Output); ok {
				messages[id] = append(messages[id], line)
			}
		}
	}

	// go test запускает пакеты параллельно и перемешивает их события, поэтому
	// порядок появления тестов в выводе от прогона к прогону разный. Отчёт
	// обязан быть воспроизводимым — сортируем по пакету и имени.
	sort.Slice(order, func(i, j int) bool {
		if order[i].pkg != order[j].pkg {
			return order[i].pkg < order[j].pkg
		}
		return order[i].test < order[j].test
	})

	suite := Suite{}
	for _, id := range order {
		result := TestResult{
			Package:  id.pkg,
			Name:     id.test,
			Status:   statuses[id],
			Messages: messages[id],
		}
		suite.Tests = append(suite.Tests, result)
		suite.Total++
		switch result.Status {
		case "pass":
			suite.Passed++
		case "fail":
			suite.Failed++
			suite.FailedNames = append(suite.FailedNames, id.pkg+"."+id.test)
		}
	}
	sort.Strings(suite.FailedNames)
	if suite.Total == 0 {
		return suite, fmt.Errorf("go test не выполнил ни одного теста")
	}
	return suite, nil
}

// assertionLine оставляет из вывода теста только строки самих проверок вида
// `    values_test.go:23: текст`. Служебные строки go test (=== RUN, --- FAIL
// и итог с временем прогона) отбрасываются: время прогона в отчёте сделало бы
// его невоспроизводимым.
func assertionLine(output string) (string, bool) {
	trimmed := strings.TrimRight(output, "\n")
	if !strings.HasPrefix(trimmed, "    ") {
		return "", false
	}
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "===") {
		return "", false
	}
	return trimmed, true
}

func shortPackage(full string) string {
	return strings.TrimPrefix(strings.TrimPrefix(full, modulePath), "/")
}

// runReference собирает временную копию проекта, подменяет в ней файлы задач
// эталонными и прогоняет те же тесты. Рабочая копия студента не меняется.
func runReference(referenceDir string, packages []string) (*Suite, error) {
	info, err := os.Stat(referenceDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("нет директории эталона %s", referenceDir)
	}
	tmp, err := os.MkdirTemp("", "gobasics-reference-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	if err := copyGoTree(".", tmp); err != nil {
		return nil, fmt.Errorf("копирование проекта: %w", err)
	}
	replaced, err := overlay(referenceDir, tmp)
	if err != nil {
		return nil, fmt.Errorf("подмена файлов эталоном: %w", err)
	}
	if replaced == 0 {
		return nil, fmt.Errorf("в %s нет ни одного файла .go", referenceDir)
	}

	suite, err := runTests(tmp, packages)
	if err != nil {
		return nil, err
	}
	return &suite, nil
}

// copyGoTree копирует go.mod и все файлы .go, сохраняя структуру директорий.
// Всё остальное (README, отчёты, случайный мусор) во временную копию не едет.
func copyGoTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if name != "go.mod" && filepath.Ext(name) != ".go" {
			return nil
		}
		relative, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		return copyFile(path, filepath.Join(to, relative))
	})
}

// overlay кладёт файлы .go из эталона поверх копии проекта по тем же
// относительным путям и возвращает число подменённых файлов.
func overlay(referenceDir, target string) (int, error) {
	count := 0
	err := filepath.WalkDir(referenceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		relative, err := filepath.Rel(referenceDir, path)
		if err != nil {
			return err
		}
		count++
		return copyFile(path, filepath.Join(target, relative))
	})
	return count, err
}

func copyFile(from, to string) error {
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	return os.WriteFile(to, data, 0o644)
}

func printSuite(label string, suite Suite) {
	fmt.Printf("\n%s: пройдено %d из %d\n", label, suite.Passed, suite.Total)
	for _, test := range suite.Tests {
		if test.Status == "pass" {
			continue
		}
		fmt.Printf("  [%s] %s.%s\n", test.Status, test.Package, test.Name)
		for _, message := range test.Messages {
			fmt.Printf("      %s\n", message)
		}
	}
}

func writeReport(path string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
