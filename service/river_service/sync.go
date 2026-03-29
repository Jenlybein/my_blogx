package river_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myblogx/global"
	"myblogx/service/es_service"
	"myblogx/service/river_service/rule"
	"reflect"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-mysql-org/go-mysql/schema"
	"github.com/pingcap/errors"
)

const (
	fieldTypeList = "list"
	// 用于mysql int类型到es日期类型的转换
	// 设置 [rule.field] created_time = ",date"
	fieldTypeDate = "date"
)

const mysqlDateFormat = "2006-01-02"

// posSaver 保存位置信息的结构
type posSaver struct {
	pos   mysql.Position // MySQL位置
	force bool           // 是否强制保存
}

// eventHandler 事件处理器
type eventHandler struct {
	r *River // River实例
}

// OnRotate 处理旋转事件
func (h *eventHandler) OnRotate(header *replication.EventHeader, e *replication.RotateEvent) error {
	pos := mysql.Position{
		Name: string(e.NextLogName),
		Pos:  uint32(e.Position),
	}

	h.r.syncCh <- posSaver{pos, true}

	return h.r.ctx.Err()
}

// OnTableChanged 处理表变更事件
func (h *eventHandler) OnTableChanged(header *replication.EventHeader, schema, table string) error {
	err := h.r.updateRule(schema, table)
	if err != nil && err != ErrRuleNotExist {
		return errors.Trace(err)
	}
	return nil
}

// OnDDL 处理DDL语句事件
func (h *eventHandler) OnDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	h.r.syncCh <- posSaver{nextPos, true}
	return h.r.ctx.Err()
}

// OnXID 处理事务提交事件
func (h *eventHandler) OnXID(header *replication.EventHeader, nextPos mysql.Position) error {
	h.r.syncCh <- posSaver{nextPos, false}
	return h.r.ctx.Err()
}

// OnRow 处理行事件
func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	rule, ok := h.r.rules[ruleKey(e.Table.Schema, e.Table.Name)]
	if !ok {
		return nil
	}

	var reqs []*es_service.BulkRequest
	var err error
	switch e.Action {
	case canal.InsertAction:
		reqs, err = h.r.makeInsertRequest(rule, e.Rows)
	case canal.DeleteAction:
		reqs, err = h.r.makeDeleteRequest(rule, e.Rows)
	case canal.UpdateAction:
		reqs, err = h.r.makeUpdateRequest(rule, e.Rows)
	default:
		err = errors.Errorf("无效的行事件操作: %s", e.Action)
	}

	if err != nil {
		h.r.cancel()
		return errors.Errorf("构建 %s 的 ES 请求失败，停止同步: %v", e.Action, err)
	}

	h.r.syncCh <- reqs

	return h.r.ctx.Err()
}

// OnGTID 处理GTID事件
func (h *eventHandler) OnGTID(header *replication.EventHeader, gtidEvent mysql.BinlogGTIDEvent) error {
	return nil
}

// OnPosSynced 处理位置同步事件
func (h *eventHandler) OnPosSynced(header *replication.EventHeader, pos mysql.Position, set mysql.GTIDSet, force bool) error {
	return nil
}

// OnRowsQueryEvent 处理行查询事件
func (h *eventHandler) OnRowsQueryEvent(e *replication.RowsQueryEvent) error {
	return nil
}

// String 返回事件处理器的字符串表示
func (h *eventHandler) String() string {
	return "ESRiverEventHandler"
}

// syncLoop 同步循环，处理同步请求
func (r *River) syncLoop() {
	bulkSize := global.Config.River.BulkSize
	if bulkSize == 0 {
		bulkSize = 128
	}

	interval := time.Duration(global.Config.River.FlushBulkTime) * time.Millisecond
	if interval == 0 {
		interval = 200 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer r.wg.Done()

	lastSavedTime := time.Now()
	reqs := make([]*es_service.BulkRequest, 0, 1024)

	var pos mysql.Position

	for {
		needFlush := false
		needSavePos := false

		select {
		case v := <-r.syncCh:
			switch v := v.(type) {
			case posSaver:
				now := time.Now()
				if v.force || now.Sub(lastSavedTime) > 3*time.Second {
					lastSavedTime = now
					needFlush = true
					needSavePos = true
					pos = v.pos
				}
			case []*es_service.BulkRequest:
				reqs = append(reqs, v...)
				needFlush = len(reqs) >= bulkSize
			}
		case <-ticker.C:
			needFlush = true
		case <-r.ctx.Done():
			return
		}

		if needFlush {
			// TODO: retry some times?
			if err := r.doBulk(reqs); err != nil {
				global.Logger.Errorf("执行 ES 批量同步失败，停止同步: %v", err)
				r.cancel()
				return
			}
			reqs = reqs[0:0]
		}

		if needSavePos {
			if err := r.master.Save(pos); err != nil {
				global.Logger.Errorf("保存同步位点失败，停止同步: 位点=%s 错误=%v", pos, err)
				r.cancel()
				return
			}
		}
	}
}

// makeRequest 为插入和删除操作创建请求
func (r *River) makeRequest(rule *rule.Rule, action string, rows [][]interface{}) ([]*es_service.BulkRequest, error) {
	reqs := make([]*es_service.BulkRequest, 0, len(rows))

	for _, values := range rows {
		id, err := r.getDocID(rule, values)
		if err != nil {
			return nil, errors.Trace(err)
		}

		parentID := ""
		if len(rule.Parent) > 0 {
			if parentID, err = r.getParentID(rule, values, rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
		}

		req := &es_service.BulkRequest{Index: rule.Index, Type: rule.Type, ID: id, Parent: parentID, Pipeline: rule.Pipeline}

		if action == canal.DeleteAction {
			req.Action = es_service.ActionDelete
		} else {
			r.makeInsertReqData(req, rule, values)
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

// makeInsertRequest 创建插入请求
func (r *River) makeInsertRequest(rule *rule.Rule, rows [][]interface{}) ([]*es_service.BulkRequest, error) {
	return r.makeRequest(rule, canal.InsertAction, rows)
}

// makeDeleteRequest 创建删除请求
func (r *River) makeDeleteRequest(rule *rule.Rule, rows [][]interface{}) ([]*es_service.BulkRequest, error) {
	return r.makeRequest(rule, canal.DeleteAction, rows)
}

// makeUpdateRequest 创建更新请求
func (r *River) makeUpdateRequest(rule *rule.Rule, rows [][]interface{}) ([]*es_service.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("更新行事件数据不完整，更新事件必须成对出现，当前行数: %d", len(rows))
	}

	reqs := make([]*es_service.BulkRequest, 0, len(rows))

	for i := 0; i < len(rows); i += 2 {
		beforeID, err := r.getDocID(rule, rows[i])
		if err != nil {
			return nil, errors.Trace(err)
		}

		afterID, err := r.getDocID(rule, rows[i+1])

		if err != nil {
			return nil, errors.Trace(err)
		}

		beforeParentID, afterParentID := "", ""
		if len(rule.Parent) > 0 {
			if beforeParentID, err = r.getParentID(rule, rows[i], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
			if afterParentID, err = r.getParentID(rule, rows[i+1], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
		}

		req := &es_service.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = es_service.ActionDelete
			reqs = append(reqs, req)

			req = &es_service.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID, Pipeline: rule.Pipeline}
			r.makeInsertReqData(req, rule, rows[i+1])

		} else {
			if len(rule.Pipeline) > 0 {
				// 管道只能在索引操作上指定
				r.makeInsertReqData(req, rule, rows[i+1])
				// 确保操作是索引，而不是创建
				req.Action = es_service.ActionIndex
				req.Pipeline = rule.Pipeline
			} else {
				r.makeUpdateReqData(req, rule, rows[i], rows[i+1])
			}
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

// makeReqColumnData 根据列类型转换数据值
func (r *River) makeReqColumnData(col *schema.TableColumn, value interface{}) interface{} {
	switch col.Type {
	case schema.TYPE_ENUM:
		switch value := value.(type) {
		case int64:
			// 对于binlog，ENUM可能是int64，但对于dump，enum是字符串
			eNum := value - 1
			if eNum < 0 || eNum >= int64(len(col.EnumValues)) {
				// 我们之前插入了无效的枚举值，所以返回空
				global.Logger.Warnf("无效的 binlog 枚举索引: 索引=%d 枚举值=%v", eNum, col.EnumValues)
				return ""
			}

			return col.EnumValues[eNum]
		}
	case schema.TYPE_SET:
		switch value := value.(type) {
		case int64:
			// 对于binlog，SET可能是int64，但对于dump，SET是字符串
			bitmask := value
			sets := make([]string, 0, len(col.SetValues))
			for i, s := range col.SetValues {
				if bitmask&int64(1<<uint(i)) > 0 {
					sets = append(sets, s)
				}
			}
			return strings.Join(sets, ",")
		}
	case schema.TYPE_BIT:
		switch value := value.(type) {
		case string:
			// 对于binlog，BIT是int64，但对于dump，BIT是字符串
			// 对于dump 0x01表示1，\0表示0
			if value == "\x01" {
				return int64(1)
			}

			return int64(0)
		}
	case schema.TYPE_STRING:
		switch value := value.(type) {
		case []byte:
			return string(value[:])
		}
	case schema.TYPE_JSON:
		var f interface{}
		var err error
		switch v := value.(type) {
		case string:
			err = json.Unmarshal([]byte(v), &f)
		case []byte:
			err = json.Unmarshal(v, &f)
		}
		if err == nil && f != nil {
			return f
		}
	case schema.TYPE_DATETIME, schema.TYPE_TIMESTAMP:
		switch v := value.(type) {
		case string:
			vt, err := time.ParseInLocation(mysql.TimeFormat, string(v), time.Local)
			if err != nil || vt.IsZero() { // 解析日期失败或零日期
				return nil
			}
			return vt.Format(time.RFC3339)
		}
	case schema.TYPE_DATE:
		switch v := value.(type) {
		case string:
			vt, err := time.Parse(mysqlDateFormat, string(v))
			if err != nil || vt.IsZero() { // 解析日期失败或零日期
				return nil
			}
			return vt.Format(mysqlDateFormat)
		}
	}

	return value
}

// getFieldParts 获取字段的部分信息
func (r *River) getFieldParts(k string, v string) (string, string, string) {
	composedField := strings.Split(v, ",")

	mysql := k
	elastic := composedField[0]
	fieldType := ""

	if 0 == len(elastic) {
		elastic = mysql
	}
	if 2 == len(composedField) {
		fieldType = composedField[1]
	}

	return mysql, elastic, fieldType
}

// makeInsertReqData 创建插入请求数据
func (r *River) makeInsertReqData(req *es_service.BulkRequest, rule *rule.Rule, values []interface{}) {
	req.Data = make(map[string]interface{}, len(values))

	req.Action = es_service.ActionIndex

	for i, c := range rule.TableInfo.Columns {
		if !rule.CheckFilter(c.Name) {
			continue
		}
		mapped := false
		for k, v := range rule.FieldMapping {
			mysql, elastic, fieldType := r.getFieldParts(k, v)
			if mysql == c.Name {
				mapped = true
				req.Data[elastic] = r.getFieldValue(&c, fieldType, values[i])
			}
		}
		if mapped == false {
			req.Data[c.Name] = r.makeReqColumnData(&c, values[i])
		}
	}
}

// makeUpdateReqData 创建更新请求数据
func (r *River) makeUpdateReqData(req *es_service.BulkRequest, rule *rule.Rule,
	beforeValues []interface{}, afterValues []interface{}) {
	req.Data = make(map[string]interface{}, len(beforeValues))

	// 如果出错可能会很危险，是否先删除？
	req.Action = es_service.ActionUpdate

	for i, c := range rule.TableInfo.Columns {
		mapped := false
		if !rule.CheckFilter(c.Name) {
			continue
		}
		if reflect.DeepEqual(beforeValues[i], afterValues[i]) {
			// 没有任何变化
			continue
		}
		for k, v := range rule.FieldMapping {
			mysql, elastic, fieldType := r.getFieldParts(k, v)
			if mysql == c.Name {
				mapped = true
				req.Data[elastic] = r.getFieldValue(&c, fieldType, afterValues[i])
			}
		}
		if mapped == false {
			req.Data[c.Name] = r.makeReqColumnData(&c, afterValues[i])
		}

	}
}

// getDocID 获取文档ID
// 如果toml文件中的id为none，则获取一行中的主键并将它们格式化为字符串，且PK不能为nil
// 否则获取一行中的ID列并将它们格式化为字符串
func (r *River) getDocID(rule *rule.Rule, row []interface{}) (string, error) {
	var (
		ids []interface{}
		err error
	)
	if rule.ID == nil {
		ids, err = rule.TableInfo.GetPKValues(row)
		if err != nil {
			return "", err
		}
	} else {
		ids = make([]interface{}, 0, len(rule.ID))
		for _, column := range rule.ID {
			value, err := rule.TableInfo.GetColumnValue(column, row)
			if err != nil {
				return "", err
			}
			ids = append(ids, value)
		}
	}

	var buf bytes.Buffer

	sep := ""
	for i, value := range ids {
		if value == nil {
			return "", errors.Errorf("The %ds id or PK value is nil", i)
		}

		buf.WriteString(fmt.Sprintf("%s%v", sep, value))
		sep = ":"
	}

	return buf.String(), nil
}

// getParentID 获取父文档ID
func (r *River) getParentID(rule *rule.Rule, row []interface{}, columnName string) (string, error) {
	index := rule.TableInfo.FindColumn(columnName)
	if index < 0 {
		return "", errors.Errorf("parent id not found %s(%s)", rule.TableInfo.Name, columnName)
	}

	return fmt.Sprint(row[index]), nil
}

// doBulk 执行批量请求
func (r *River) doBulk(reqs []*es_service.BulkRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	resp := es_service.Bulk(reqs)

	errorsCount := 0
	successCount := 0

	// 输出响应数据
	if resp.Data != nil {
		dataBytes, err := json.Marshal(resp.Data)
		if err == nil {
			// 解析响应数据，分析每个操作的状态
			var bulkResp map[string]interface{}
			if json.Unmarshal(dataBytes, &bulkResp) == nil {
				if items, ok := bulkResp["items"].([]interface{}); ok {
					for _, item := range items {
						if itemMap, ok := item.(map[string]interface{}); ok {
							for action, result := range itemMap {
								if resultMap, ok := result.(map[string]interface{}); ok {
									status := int(resultMap["status"].(float64))
									var resultStr string
									if resultVal, exists := resultMap["result"]; exists && resultVal != nil {
										resultStr = resultVal.(string)
									}
									docID := resultMap["_id"].(string)

									// 检查是否有错误信息
									var errorMsg string
									if errorVal, exists := resultMap["error"]; exists && errorVal != nil {
										if errorMap, ok := errorVal.(map[string]interface{}); ok {
											if reason, exists := errorMap["reason"].(string); exists {
												errorMsg = reason
											}
											if causedBy, exists := errorMap["caused_by"].(map[string]interface{}); exists {
												if cbReason, exists := causedBy["reason"].(string); exists {
													errorMsg = errorMsg + " (原因: " + cbReason + ")"
												}
											}
										}
									}

									// 根据状态和结果分析
									switch {
									case status >= 200 && status < 300:
										successCount++
									default:
										errorsCount++
										if errorMsg != "" {
											global.Logger.Errorf("同步错误: %s 操作文档 %s 失败，状态: %d, 错误: %s", action, docID, status, errorMsg)
										} else {
											global.Logger.Errorf("同步错误: %s 操作文档 %s 失败，状态: %d, 结果: %s", action, docID, status, resultStr)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 输出操作统计信息
	global.Logger.Infof("同步操作统计: 总计=%d, 成功=%d, 失败=%d, binlog=%s",
		len(reqs), successCount, errorsCount, r.canal.SyncedPosition())

	if !resp.Success {
		global.Logger.Errorf("同步文档失败 %s, binlog=%s", resp.Msg, r.canal.SyncedPosition())
	}

	return nil
}

// getFieldValue 获取mysql字段值并将其转换为特定的es值
func (r *River) getFieldValue(col *schema.TableColumn, fieldType string, value interface{}) interface{} {
	var fieldValue interface{}
	switch fieldType {
	case fieldTypeList:
		v := r.makeReqColumnData(col, value)
		if str, ok := v.(string); ok {
			fieldValue = strings.Split(str, ",")
		} else {
			fieldValue = v
		}

	case fieldTypeDate:
		if col.Type == schema.TYPE_NUMBER {
			col.Type = schema.TYPE_DATETIME

			v := reflect.ValueOf(value)
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fieldValue = r.makeReqColumnData(col, time.Unix(v.Int(), 0).Format(mysql.TimeFormat))
			}
		}
	}

	if fieldValue == nil {
		fieldValue = r.makeReqColumnData(col, value)
	}
	return fieldValue
}
