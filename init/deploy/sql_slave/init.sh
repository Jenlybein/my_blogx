#!/bin/bash
set -e

# 等待主库启动完成
echo "等待主库 mysql-master (10.2.0.2:3306) 启动..."
until mysql -h mysql-master -uroot -pREDACTED_SHARED_PASSWORD -e "SELECT 1" > /dev/null 2>&1; do
  sleep 2
done

# 从主库获取binlog文件和位置（关键修正：先查询，再赋值）
echo "获取主库binlog信息..."
MASTER_LOG_FILE=$(mysql -h mysql-master -uroot -pREDACTED_SHARED_PASSWORD -e "SHOW MASTER STATUS\G" | grep "File:" | awk '{print $2}')
MASTER_LOG_POS=$(mysql -h mysql-master -uroot -pREDACTED_SHARED_PASSWORD -e "SHOW MASTER STATUS\G" | grep "Position:" | awk '{print $2}')

echo "主库binlog文件：$MASTER_LOG_FILE，位置：$MASTER_LOG_POS"

# 从库初始化配置
echo "开始配置从库..."
mysql -uroot -pREDACTED_SHARED_PASSWORD << EOF
-- 1. 设置从库server_id（确保唯一）
SET GLOBAL server_id = 2;
-- 2. 关闭只读（临时）
SET GLOBAL read_only = 0;
-- 3. 重置从库（避免旧配置干扰）
RESET SLAVE ALL;
-- 4. 配置主从复制（使用上面获取的binlog信息）
CHANGE MASTER TO
MASTER_HOST='mysql-master',
MASTER_USER='repl',
MASTER_PASSWORD='REDACTED_SHARED_PASSWORD',
MASTER_PORT=3306,
MASTER_LOG_FILE='$MASTER_LOG_FILE',
MASTER_LOG_POS=$MASTER_LOG_POS,
MASTER_CONNECT_RETRY=10;
-- 5. 启动从库复制进程
START SLAVE;
-- 6. 开启只读（仅允许super权限）
SET GLOBAL read_only = 1;
-- 7. 查看从库状态
SHOW SLAVE STATUS\G;
EOF

# 创建数据库 blogx
# 优化后（推荐）
mysql -h mysql-master -uroot -pREDACTED_SHARED_PASSWORD -e "CREATE DATABASE IF NOT EXISTS blogx DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"

echo "从库配置完成！"
