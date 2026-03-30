-- 创建 blogx_logs 数据库（不存在则创建），用于统一存储各类日志数据
CREATE DATABASE IF NOT EXISTS blogx_logs;

-- ==============================================
-- 运行时日志表：存储系统通用运行日志、接口请求日志
-- ==============================================
CREATE TABLE IF NOT EXISTS blogx_logs.runtime_logs (
    event_id UInt64, -- 日志事件唯一ID
    ts DateTime64 (3, 'Asia/Shanghai'), -- 日志时间戳（毫秒精度，上海时区）
    log_kind LowCardinality (String), -- 日志类型（低基数优化，如 runtime/api）
    service LowCardinality (String), -- 服务名称（低基数优化）
    env LowCardinality (String), -- 环境标识（dev/test/prod）
    host LowCardinality (String), -- 服务器主机名/IP
    instance_id LowCardinality (String), -- 服务实例ID
    level LowCardinality (String), -- 日志级别（info/warn/error）
    message String, -- 日志描述信息
    request_id String, -- 请求ID（用于链路追踪）
    trace_id String, -- 链路追踪ID
    file String, -- 日志输出代码文件
    func String, -- 日志输出函数/方法
    user_id UInt64, -- 用户ID（无则为0）
    ip String, -- 用户客户端IP
    method LowCardinality (String), -- HTTP请求方法（GET/POST/PUT等）
    path String, -- HTTP请求路径
    status_code UInt16, -- HTTP响应状态码
    latency_ms UInt32, -- 请求耗时（毫秒）
    event_name LowCardinality (String), -- 事件名称
    error_type String, -- 错误类型（异常类名）
    error_stack String, -- 错误堆栈信息
    extra_json String -- 扩展字段（JSON格式）
) ENGINE = MergeTree -- 表引擎：MergeTree（ClickHouse核心引擎）
PARTITION BY
    toYYYYMM (ts) -- 分区规则：按日志时间的年月分区
ORDER BY ( -- 排序键：查询优化索引
        service,
        ts,
        level,
        host,
        event_id
    ) TTL toDateTime (ts) + INTERVAL 90 DAY -- 数据自动过期：保留90天
    SETTINGS index_granularity = 8192;
-- 索引粒度参数（默认优化值）

-- ==============================================
-- 登录事件日志表：专门存储用户登录/登出相关日志
-- ==============================================
CREATE TABLE IF NOT EXISTS blogx_logs.login_event_logs (
    event_id UInt64, -- 日志事件唯一ID
    ts DateTime64 (3, 'Asia/Shanghai'), -- 日志时间戳（毫秒精度，上海时区）
    log_kind LowCardinality (String), -- 日志类型
    service LowCardinality (String), -- 服务名称
    env LowCardinality (String), -- 环境标识
    host LowCardinality (String), -- 服务器主机
    instance_id LowCardinality (String), -- 服务实例ID
    level LowCardinality (String), -- 日志级别
    message String, -- 日志描述信息
    request_id String, -- 请求ID
    trace_id String, -- 链路追踪ID
    user_id UInt64, -- 用户ID
    ip String, -- 登录IP
    event_name LowCardinality (String), -- 事件名称（login/logout）
    username String, -- 登录账号/用户名
    login_type LowCardinality (String), -- 登录类型（password/sms/oidc等）
    success UInt8, -- 是否成功（1=成功，0=失败）
    reason String, -- 失败原因/登录结果描述
    addr String, -- IP归属地/地理位置
    ua String, -- 浏览器/客户端UA信息
    extra_json String -- 扩展字段（JSON格式）
) ENGINE = MergeTree
PARTITION BY
    toYYYYMM (ts) -- 按年月分区
ORDER BY ( -- 排序索引
        user_id,
        ts,
        login_type,
        success,
        event_id
    ) TTL toDateTime (ts) + INTERVAL 365 DAY -- 数据保留1年
    SETTINGS index_granularity = 8192;

-- ==============================================
-- 操作审计日志表：存储用户关键行为审计日志
-- ==============================================
CREATE TABLE IF NOT EXISTS blogx_logs.action_audit_logs (
    event_id UInt64, -- 日志事件唯一ID
    ts DateTime64 (3, 'Asia/Shanghai'), -- 日志时间戳（毫秒精度，上海时区）
    log_kind LowCardinality (String), -- 日志类型
    service LowCardinality (String), -- 服务名称
    env LowCardinality (String), -- 环境标识
    host LowCardinality (String), -- 服务器主机
    instance_id LowCardinality (String), -- 服务实例ID
    level LowCardinality (String), -- 日志级别
    message String, -- 日志描述信息
    request_id String, -- 请求ID
    trace_id String, -- 链路追踪ID
    user_id UInt64, -- 操作用户ID
    ip String, -- 操作IP
    method LowCardinality (String), -- 请求方法
    path String, -- 请求接口路径
    status_code UInt16, -- 响应状态码
    action_name LowCardinality (String), -- 操作名称（如 create/update/delete）
    target_type LowCardinality (String), -- 操作目标类型（如 user/article/comment）
    target_id String, -- 操作目标ID
    success UInt8, -- 操作是否成功（1=成功，0=失败）
    request_body String, -- 请求体内容
    response_body String, -- 响应体内容
    request_body_raw String, -- 脱敏截断后的原始请求体
    response_body_raw String, -- 脱敏截断后的原始响应体
    request_header_raw String, -- 脱敏截断后的原始请求头
    response_header_raw String, -- 脱敏截断后的原始响应头
    extra_json String -- 扩展字段（JSON格式）
) ENGINE = MergeTree
PARTITION BY
    toYYYYMM (ts) -- 按年月分区
ORDER BY ( -- 排序索引
        user_id,
        ts,
        action_name,
        target_type,
        event_id
    ) TTL toDateTime (ts) + INTERVAL 365 DAY -- 数据保留1年
    SETTINGS index_granularity = 8192;

-- 为已存在的操作审计日志表补齐原始请求/响应体字段，避免老环境升级后入库失败。
ALTER TABLE blogx_logs.action_audit_logs
    ADD COLUMN IF NOT EXISTS request_body_raw String AFTER response_body;

ALTER TABLE blogx_logs.action_audit_logs
    ADD COLUMN IF NOT EXISTS response_body_raw String AFTER request_body_raw;

ALTER TABLE blogx_logs.action_audit_logs
    ADD COLUMN IF NOT EXISTS request_header_raw String AFTER response_body_raw;

ALTER TABLE blogx_logs.action_audit_logs
    ADD COLUMN IF NOT EXISTS response_header_raw String AFTER request_header_raw;
