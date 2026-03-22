#!/bin/sh
set -e

# 插件已安装时直接跳过，避免容器重启后重复安装失败
if [ -d /usr/share/elasticsearch/plugins/analysis-ik ] || \
   /usr/share/elasticsearch/bin/elasticsearch-plugin list 2>/dev/null | grep -qx "analysis-ik"; then
  echo "analysis-ik 插件已存在，跳过安装"
else
  /usr/share/elasticsearch/bin/elasticsearch-plugin install -b https://get.infini.cloud/elasticsearch/analysis-ik/7.12.0
fi

# 启动 ES（本次启动就会加载新插件）
exec /usr/local/bin/docker-entrypoint.sh elasticsearch
