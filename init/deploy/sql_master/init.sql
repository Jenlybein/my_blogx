-- TODO：后续需用 mysqldump 导出完整数据库，替换这里的操作。263 集

-- 创建新用户 admin，允许从任意IP访问
CREATE USER 'admin' @'%' IDENTIFIED BY 'REDACTED_SHARED_PASSWORD';

-- 授予 admin 用户全部权限（和 root 等价）
GRANT ALL PRIVILEGES ON *.* TO 'admin' @'%' WITH GRANT OPTION;

-- 授予复制权限：允许任何ip通过登录用户 repl 来获取复制权限，密码 REDACTED_SHARED_PASSWORD
GRANT REPLICATION SLAVE ON *.* TO 'repl' @'%' identified by 'REDACTED_SHARED_PASSWORD';

-- 刷新权限
FLUSH PRIVILEGES;