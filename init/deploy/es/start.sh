#!/bin/sh
set -e

# 安装插件
if [ ! -d /usr/share/elasticsearch/plugins/analysis-ik/config ]; then
  /usr/share/elasticsearch/bin/elasticsearch-plugin install -b https://get.infini.cloud/elasticsearch/analysis-ik/7.12.0
fi

# 启动 ES（本次启动就会加载新插件）
exec /usr/local/bin/docker-entrypoint.sh elasticsearch