# 系统架构基础示意：
flowchart TB
    %% 客户端层
    subgraph C["客户端层"]
        U["普通用户前端"]
        A["后台管理前端"]
    end

    %% 接入层
    subgraph G["接入层"]
        N["Nginx / HTTPS / 反向代理"]
        WS["WebSocket 私信入口"]
        SSE["SSE AI 流式响应"]
    end

    %% 应用层
    subgraph APP["应用层 Gin 后端"]
        R["路由层"]
        M["中间件层"]
        API["API 处理层"]
        SVC["业务服务层"]
        CORE["核心全局组件"]
    end

    %% 核心业务模块
    subgraph BIZ["核心业务模块"]
        AUTH["用户认证|关系|主页"]
        ARTICLE["文章管理"]
        IMAGE["图片资源"]
        COMMENT["评论互动"]
        MESSAGE["消息|私信"]
        SEARCH["全文搜索"]
        AI["AI 模块"]
        DATA["数据统计"]
        LOG["日志审计"]
        SITE["站点配置"]
    end
    %% 业务模块换行处理，避免太长
        MESSAGE ~~~ COMMENT
        AUTH ~~~ LOG
        SEARCH ~~~ AI
        SITE ~~~ ARTICLE
        DATA ~~~ IMAGE

    %% 数据与基础设施
    subgraph DB["数据与存储层"]
        REDIS["Redis 缓存"]
        ES["Elasticsearch 搜索"]
        CH["ClickHouse 日志库"]
        MYSQL["MySQL 业务库"]
        QN["七牛云对象存储"]
    end

    %% 异步与外部服务
    subgraph EXT["异步与外部协同"]
        CRON["定时任务"]
        RIVER["图片引用监听"]
        LLM["大模型服务"]
        FB["日志采集"]
    end

    %% 主流程：
    C --> G
    G --> APP
    APP --> BIZ
    BIZ ~~~ DB
    DB --> EXT
    EXT ~~~ BIZ

    %% 精简关键关联
    CORE --> CRON
    MYSQL -.监听.-> RIVER
    RIVER --> MYSQL
    LOG --> FB
    FB --> CH
    AI --> LLM
    QN -.回调.-> IMAGE
    ARTICLE -.同步.-> ES
    CRON -.回刷.-> MYSQL



