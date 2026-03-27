package image_ref_river_service

import (
	"fmt"
	"strings"
	"time"

	"myblogx/models/ctype"
)

type rowSnapshot struct {
	columns map[string]int
	row     []any
}

func newRowSnapshot(columnNames []string, row []any) rowSnapshot {
	columns := make(map[string]int, len(columnNames))
	for index, name := range columnNames {
		columns[name] = index
	}
	return rowSnapshot{
		columns: columns,
		row:     row,
	}
}

func (r rowSnapshot) ID() (ctype.ID, error) {
	value, ok := r.value("id")
	if !ok {
		return 0, fmt.Errorf("canal 行数据缺少 id 列")
	}
	var id ctype.ID
	if err := id.Scan(value); err != nil || id == 0 {
		return 0, fmt.Errorf("canal 行数据 id 解析失败")
	}
	return id, nil
}

func (r rowSnapshot) RequireString(column string) (string, error) {
	value, ok := r.value(column)
	if !ok {
		return "", fmt.Errorf("canal 行数据缺少 %s 列，请确认 binlog_row_image=FULL", column)
	}
	return normalizeString(value), nil
}

func (r rowSnapshot) IsDeleted() bool {
	value, ok := r.value("deleted_at")
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case nil:
		return false
	case time.Time:
		return !typed.IsZero()
	case *time.Time:
		return typed != nil && !typed.IsZero()
	case []byte:
		return strings.TrimSpace(string(typed)) != ""
	case string:
		return strings.TrimSpace(typed) != ""
	default:
		return fmt.Sprint(typed) != ""
	}
}

func (r rowSnapshot) value(column string) (any, bool) {
	index, ok := r.columns[column]
	if !ok || index >= len(r.row) {
		return nil, false
	}
	return r.row[index], true
}

func normalizeString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(typed)
	}
}
