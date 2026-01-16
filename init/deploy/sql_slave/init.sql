-- 1. 创建新用户 admin，允许从任意IP访问
CREATE USER 'admin' @'%' IDENTIFIED BY 'REDACTED_SHARED_PASSWORD';

-- 2. 授予 admin 用户全部权限（和 root 等价）
GRANT ALL PRIVILEGES ON *.* TO 'admin' @'%' WITH GRANT OPTION;

-- 3. 刷新权限
FLUSH PRIVILEGES;