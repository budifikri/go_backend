package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type CRUDCount struct {
	Create int `json:"create"`
	Update int `json:"update"`
	Delete int `json:"delete"`
	Error  int `json:"error"`
}

type Logger struct {
	logDir     string
	enableCRUD bool
	mu         sync.Mutex
	counters   map[string]CRUDCount
}

func NewLogger(logDir string, enableCRUD bool) *Logger {
	if logDir == "" {
		logDir = "logs"
	}
	l := &Logger{
		logDir:     logDir,
		enableCRUD: enableCRUD,
		counters:   make(map[string]CRUDCount),
	}
	_ = os.MkdirAll(l.currentMonthDir(), 0o755)
	return l
}

func (l *Logger) Log(action, table, userID, companyID, recordID string, payload interface{}) {
	if !l.enableCRUD {
		return
	}
	table = sanitizeTableName(table)
	action = strings.ToUpper(action)

	line := fmt.Sprintf("[%s] [INFO] [%s] [%s] user_id=%s company_id=%s", time.Now().Format("2006-01-02 15:04:05"), action, table, defaultValue(userID), defaultValue(companyID))
	if action == "CREATE" {
		line += " data=" + normalizePayload(payload)
	} else {
		line += " record_id=" + defaultValue(recordID)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_ = os.MkdirAll(l.currentMonthDir(), 0o755)
	_ = appendLine(l.tableLogPath(table), line)

	count := l.counters[table]
	switch strings.ToUpper(action) {
	case "CREATE":
		count.Create++
	case "UPDATE":
		count.Update++
	case "DELETE":
		count.Delete++
	}
	l.counters[table] = count
}

func (l *Logger) LogError(action, table, userID, companyID, recordID string, err error) {
	if !l.enableCRUD {
		return
	}
	table = sanitizeTableName(table)
	errMsg := "unknown error"
	if err != nil {
		errMsg = err.Error()
	}
	line := fmt.Sprintf(
		"[%s] [ERROR] [%s] [%s] user_id=%s company_id=%s record_id=%s error=%s",
		time.Now().Format("2006-01-02 15:04:05"),
		strings.ToUpper(action),
		table,
		defaultValue(userID),
		defaultValue(companyID),
		defaultValue(recordID),
		errMsg,
	)

	l.mu.Lock()
	defer l.mu.Unlock()

	_ = os.MkdirAll(l.currentMonthDir(), 0o755)
	_ = appendLine(l.errorLogPath(), line)

	count := l.counters[table]
	count.Error++
	l.counters[table] = count
}

func (l *Logger) GetSummary() map[string]CRUDCount {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := make(map[string]CRUDCount, len(l.counters))
	for k, v := range l.counters {
		result[k] = v
	}
	return result
}

func (l *Logger) SaveSummary() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enableCRUD {
		return nil
	}

	monthDir := l.currentMonthDir()
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		return err
	}

	var b strings.Builder
	monthName := time.Now().Format("2006_01")
	b.WriteString("=== CRUD Summary - ")
	b.WriteString(monthName)
	b.WriteString(" ===\n\n")
	b.WriteString("Table          | CREATE | UPDATE | DELETE | ERROR\n")
	b.WriteString("---------------|--------|--------|--------|-------\n")

	tables := make([]string, 0, len(l.counters))
	for table := range l.counters {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	totalCreate := 0
	totalUpdate := 0
	totalDelete := 0
	totalError := 0

	for _, table := range tables {
		c := l.counters[table]
		totalCreate += c.Create
		totalUpdate += c.Update
		totalDelete += c.Delete
		totalError += c.Error
		b.WriteString(fmt.Sprintf("%-14s | %6d | %6d | %6d | %5d\n", table, c.Create, c.Update, c.Delete, c.Error))
	}

	b.WriteString("---------------|--------|--------|--------|-------\n")
	b.WriteString(fmt.Sprintf("%-14s | %6d | %6d | %6d | %5d\n", "TOTAL", totalCreate, totalUpdate, totalDelete, totalError))

	return os.WriteFile(filepath.Join(monthDir, "summary.txt"), []byte(b.String()), 0o644)
}

func (l *Logger) ListFiles() (map[string][]string, error) {
	result := make(map[string][]string)
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		monthPath := filepath.Join(l.logDir, entry.Name())
		monthEntries, readErr := os.ReadDir(monthPath)
		if readErr != nil {
			continue
		}
		files := make([]string, 0)
		for _, monthEntry := range monthEntries {
			if monthEntry.IsDir() {
				continue
			}
			files = append(files, monthEntry.Name())
		}
		sort.Strings(files)
		result[entry.Name()] = files
	}

	return result, nil
}

func (l *Logger) ReadTableLogs(yearMonth, table string, limit, offset int) ([]string, int64, error) {
	table = sanitizeTableName(table)
	filePath := filepath.Join(l.logDir, yearMonth, table+".log")
	return readLinesWithPagination(filePath, limit, offset)
}

func (l *Logger) ReadErrorLogs(yearMonth string, limit, offset int) ([]string, int64, error) {
	filePath := filepath.Join(l.logDir, yearMonth, "error.log")
	return readLinesWithPagination(filePath, limit, offset)
}

func (l *Logger) currentMonthDir() string {
	return filepath.Join(l.logDir, time.Now().Format("2006_01"))
}

func (l *Logger) tableLogPath(table string) string {
	return filepath.Join(l.currentMonthDir(), table+".log")
}

func (l *Logger) errorLogPath() string {
	return filepath.Join(l.currentMonthDir(), "error.log")
}

func sanitizeTableName(table string) string {
	table = strings.TrimSpace(strings.ToLower(table))
	if table == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, ch := range table {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			b.WriteRune(ch)
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

func defaultValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func normalizePayload(payload interface{}) string {
	if payload == nil {
		return "{}"
	}

	switch v := payload.(type) {
	case []byte:
		return normalizeJSONBytes(v)
	case string:
		return normalizeJSONBytes([]byte(v))
	default:
		b, err := json.Marshal(v)
		if err != nil {
			fallback, _ := json.Marshal(fmt.Sprintf("%v", v))
			return string(fallback)
		}
		return string(b)
	}
}

func normalizeJSONBytes(b []byte) string {
	trimmed := strings.TrimSpace(string(b))
	if trimmed == "" {
		return "{}"
	}

	var js interface{}
	if err := json.Unmarshal([]byte(trimmed), &js); err != nil {
		fallback, _ := json.Marshal(trimmed)
		return string(fallback)
	}
	normalized, err := json.Marshal(js)
	if err != nil {
		fallback, _ := json.Marshal(trimmed)
		return string(fallback)
	}
	return string(normalized)
}

func appendLine(filePath, line string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

func readLinesWithPagination(filePath string, limit, offset int) ([]string, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, 0, nil
		}
		return nil, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	total := int64(len(lines))
	if offset >= len(lines) {
		return []string{}, total, nil
	}
	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}

	return lines[offset:end], total, nil
}
