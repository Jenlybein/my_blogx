package image_ref_river_service

import (
	"strings"

	"myblogx/global"
	"myblogx/models/ctype"
	"myblogx/models/enum/image_ref_enum"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

// River 基于 MySQL Binlog 的图片引用关系监听服务
// 作用：监听文章/用户/横幅/收藏表的数据变化，自动维护图片引用关系
type River struct {
	canal *canal.Canal
}

// NewRiver 初始化图片引用关系监听服务
func NewRiver() (*River, error) {
	r := &River{}
	// 初始化 canal 客户端
	if err := r.newCanal(); err != nil {
		return nil, err
	}
	// 设置 binlog 事件处理器
	r.canal.SetEventHandler(&eventHandler{})
	return r, nil
}

// newCanal 配置并创建 MySQL Binlog 监听客户端
func (r *River) newCanal() error {
	// 创建默认 canal 配置
	cfg := canal.NewDefaultConfig()
	// 绑定日志适配器
	cfg.Logger = logrusToSlogAdapter(global.Logger)
	// 从全局配置加载 MySQL 连接信息
	cfg.Addr = global.Config.ImageRefRiver.Mysql.Addr
	cfg.User = global.Config.ImageRefRiver.Mysql.User
	cfg.Password = global.Config.ImageRefRiver.Mysql.Password
	cfg.Charset = global.Config.ImageRefRiver.Charset
	cfg.Flavor = global.Config.ImageRefRiver.Flavor
	cfg.ServerID = global.Config.ImageRefRiver.ServerID
	// 全量数据备份相关配置
	cfg.Dump.ExecutionPath = global.Config.ImageRefRiver.DumpExec
	cfg.Dump.SkipMasterData = global.Config.ImageRefRiver.SkipMasterData

	// 获取要监听的数据库名
	schema := strings.TrimSpace(global.Config.ImageRefRiver.Schema)
	// 监听四张核心业务表
	for _, table := range []string{"article_models", "user_models", "banner_models", "favorite_models"} {
		cfg.IncludeTableRegex = append(cfg.IncludeTableRegex, schema+"\\."+table)
	}

	var err error
	// 创建 canal 实例
	r.canal, err = canal.NewCanal(cfg)
	if err != nil {
		return err
	}
	// 添加需要全量转储的表
	r.canal.AddDumpTables(schema, "article_models", "user_models", "banner_models", "favorite_models")
	return nil
}

// Run 启动图片引用监听服务
func (r *River) Run() error {
	return r.canal.Run()
}

// eventHandler 实现 canal.EventHandler 接口，处理数据库行事件
type eventHandler struct{}

// OnRow 处理数据表行变化（增/删/改）
func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	// 根据表名获取对应的引用类型和重建函数
	refType, rebuildByRow, ok := tableHandler(e.Table.Name)
	if !ok {
		// 非监听表，直接忽略
		return nil
	}

	// 当前批次行的列布局固定，索引映射只构建一次并在本批次复用。
	columnNames := make([]string, 0, len(e.Table.Columns))
	for _, column := range e.Table.Columns {
		columnNames = append(columnNames, column.Name)
	}
	layout := newRowLayout(columnNames)

	// 根据操作类型处理
	switch e.Action {
	case canal.DeleteAction:
		// 删除操作：删除该业务对象关联的所有图片引用
		for _, ownerID := range extractRowIDs(e) {
			if err := DeleteOwnerRefs(global.DB, refType, ownerID); err != nil {
				return err
			}
		}
	case canal.InsertAction:
		// 新增操作：重建该条数据的图片引用关系
		for _, row := range e.Rows {
			if err := rebuildByRow(newRowSnapshot(layout, row)); err != nil {
				return err
			}
		}
	case canal.UpdateAction:
		// 更新操作：使用新数据重建图片引用（e.Rows[i] 是更新后的数据）
		for i := 1; i < len(e.Rows); i += 2 {
			before := newRowSnapshot(layout, e.Rows[i-1])
			after := newRowSnapshot(layout, e.Rows[i])
			if !shouldRebuildOnUpdate(e.Table.Name, before, after) {
				continue
			}
			if err := rebuildByRow(after); err != nil {
				return err
			}
		}
	}
	return nil
}

// 以下方法为 canal.EventHandler 接口必须实现的空实现
func (h *eventHandler) OnRotate(_ *replication.EventHeader, _ *replication.RotateEvent) error {
	return nil
}
func (h *eventHandler) OnTableChanged(_ *replication.EventHeader, _, _ string) error { return nil }
func (h *eventHandler) OnDDL(_ *replication.EventHeader, _ mysql.Position, _ *replication.QueryEvent) error {
	return nil
}
func (h *eventHandler) OnXID(_ *replication.EventHeader, _ mysql.Position) error { return nil }
func (h *eventHandler) OnGTID(_ *replication.EventHeader, _ mysql.BinlogGTIDEvent) error {
	return nil
}
func (h *eventHandler) OnPosSynced(_ *replication.EventHeader, _ mysql.Position, _ mysql.GTIDSet, _ bool) error {
	return nil
}
func (h *eventHandler) OnRowsQueryEvent(_ *replication.RowsQueryEvent) error { return nil }

// String 返回事件处理器名称
func (h *eventHandler) String() string {
	return "ImageRefRiverEventHandler"
}

// tableHandler 根据表名映射：图片引用类型 + 对应行数据处理函数
func tableHandler(table string) (image_ref_enum.RefType, func(rowSnapshot) error, bool) {
	switch table {
	case "article_models":
		return image_ref_enum.RefTypeArticle, RebuildArticleRefsByRow, true
	case "user_models":
		return image_ref_enum.RefTypeUser, RebuildUserRefsByRow, true
	case "banner_models":
		return image_ref_enum.RefTypeBanner, RebuildBannerRefsByRow, true
	case "favorite_models":
		return image_ref_enum.RefTypeFavorite, RebuildFavoriteRefsByRow, true
	default:
		return 0, nil, false
	}
}

func shouldRebuildOnUpdate(table string, before rowSnapshot, after rowSnapshot) bool {
	beforeDeleted := before.IsDeleted()
	afterDeleted := after.IsDeleted()
	if beforeDeleted != afterDeleted {
		return true
	}
	if afterDeleted {
		return false
	}
	switch table {
	case "article_models":
		return !before.EqualString(after, "content") || !before.EqualString(after, "cover")
	case "user_models":
		return !before.EqualString(after, "avatar")
	case "banner_models":
		return !before.EqualString(after, "cover")
	case "favorite_models":
		return !before.EqualString(after, "cover")
	default:
		return true
	}
}

// extractRowIDs 从 binlog 行数据中提取所有主键 ID（去重）
func extractRowIDs(e *canal.RowsEvent) []ctype.ID {
	// 找到 id 列的索引位置
	index := findIDColumnIndex(e)
	if index < 0 {
		return nil
	}

	result := make([]ctype.ID, 0, len(e.Rows))
	seen := make(map[ctype.ID]struct{}, len(e.Rows)) // 去重

	// 遍历每一行数据，提取 ID
	for _, row := range e.Rows {
		id, ok := rowValueToID(row[index])
		if !ok {
			continue
		}
		// 已存在则跳过
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// findIDColumnIndex 查找表中 id 列的索引
func findIDColumnIndex(e *canal.RowsEvent) int {
	for index, column := range e.Table.Columns {
		if column.Name == "id" {
			return index
		}
	}
	return -1
}

// rowValueToID 将 binlog 中的字段值转换为业务 ID 类型
func rowValueToID(value any) (ctype.ID, bool) {
	var id ctype.ID
	// 扫描值到自定义 ID 类型
	if err := id.Scan(value); err != nil || id == 0 {
		return 0, false
	}
	return id, true
}
