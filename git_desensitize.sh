#!/bin/bash
# ==============================================================
# Git 配置文件全自动脱敏脚本
# 使用说明：1.修改顶部【配置区】的files和fields 
#          2.执行脚本：bash git_desensitize.sh
# 核心特性：一键生成clean/smudge过滤器 + 绑定.gitattributes
# ==============================================================

# ========== 【唯一配置区】 ==========
# 要脱敏的文件列表
files=(
  "settings.yaml"
  "site_default_settings.yaml"
  # "src/config/.env"
  # "application.yml"
)

# 要脱敏的字段名列表
fields=(
  "password"
  "secret"
  "issuer"
  "addr"
  "host"
  "port"
  "ip"
  "domain"
  "send_email"
  "app_key"
  "access_key"
  "secret_key"
  "addresses"
  "base_url"
)

# ========== 【常量定义】 ==========
FILTER_NAME="auto_sensitive_filter"          # Git过滤器名称
GIT_ATTRIBUTES=".gitattributes"              # 绑定文件名称
BASE_PLACEHOLDER="******"                    # 占位符基础前缀

# ========== 【0.前置一键校验】 ==========
check_env(){
  [ ! -d ".git" ] && echo -e "\033[31m请在Git项目根目录执行！\033[0m" && exit 1
  [ ! $(command -v git) ] && echo -e "\033[31m未检测到Git环境！\033[0m" && exit 1
  [ ! $(command -v sed) ] && echo -e "\033[31m未检测到sed命令！\033[0m" && exit 1
} && check_env

# ========== 【1.扫描提取敏感值+去重+生成映射】 ==========
echo -e "\n\033[36m===== 扫描目标文件，提取敏感值 =====\033[0m"
declare -A FIELD_COUNTER                     # 字段计数器：记录每个字段有效序号
declare -a VALUE_MAPPING                     # 核心映射表：存储 原始值|字段名|序号
declare -A DUPLICATE_CHECK_MAP               # 去重校验MAP：key=字段名_原始值，实现双重去重

# 初始化字段计数器
for field in "${fields[@]}"; do
  FIELD_COUNTER[$field]=0
done

# 遍历所有目标文件
for file in "${files[@]}"; do
  if [ ! -f "$file" ]; then
    echo -e "\033[33m文件 $file 不存在，跳过\033[0m"
    continue
  fi

  # 遍历所有脱敏字段
  for field in "${fields[@]}"; do
    all_origin_values=$(sed -n "s/^[[:space:]]*${field}[[:space:]]*[:=][[:space:]]*[\"']*\(.*\)[\"']*[[:space:]]*$/\1/p" "$file")
    
    # 遍历所有提取到的原始值
    while IFS= read -r origin_value; do
      if [ -z "$origin_value" ]; then
        continue
      fi
      
      # 去重：字段名+原始值 作为KEY，完全一致则判定为重复
      unique_check_key="${field}_${origin_value}"
      if [[ -n "${DUPLICATE_CHECK_MAP[$unique_check_key]}" ]]; then
        echo -e "\033[33m重复：$file -> $field = $origin_value\033[0m"
        continue
      fi

      # 非重复数据：序号自增 + 标记去重KEY + 存入映射表
      ((FIELD_COUNTER[$field]++))
      seq_num=${FIELD_COUNTER[$field]}
      DUPLICATE_CHECK_MAP[$unique_check_key]=1
      VALUE_MAPPING+=("$origin_value|$field|$seq_num")
      echo -e "\033[32m提取：$file -> $field$seq_num = $origin_value\033[0m"
    done <<< "$all_origin_values"
  done
done

# 校验是否成功获取到原始值
if [ ${#VALUE_MAPPING[@]} -eq 0 ]; then
  echo -e "\033[31m错误：未扫描到任何有效字段值\033[0m"
  exit 1
fi

# ========== 【2.生成clean/smudge规则】 ==========
echo -e "\n\033[36m===== 生成clean/smudge规则 =====\033[0m"
CLEAN_RULE=""
SMUDGE_RULE=""

for item in "${VALUE_MAPPING[@]}"; do
  # 解析映射项：原始值|字段名|序号
  val=$(echo "$item" | cut -d'|' -f1)
  field=$(echo "$item" | cut -d'|' -f2)
  num=$(echo "$item" | cut -d'|' -f3)
  placeholder="${BASE_PLACEHOLDER}_${field}${num}"
  
  # 转义所有特殊字符
  val_esc=$(echo "$val" | sed 's/[\/\*\$&|]/\\&/g')
  ph_esc=$(echo "$placeholder" | sed 's/[\/\*\$&|]/\\&/g')
  
  # 生成sed规则：将原始值替换为占位符
  CLEAN_RULE+=" -e 's/${val_esc}/${ph_esc}/g'"
  SMUDGE_RULE+=" -e 's/${ph_esc}/${val_esc}/g'"
done

echo -e "\033[32m生成${#VALUE_MAPPING[@]}条\033[0m"

# ========== 【3.配置Git全局过滤器】 ==========
echo -e "\n\033[36m===== 配置Git全局过滤器 =====\033[0m"
# 发起配置命令
git config filter.$FILTER_NAME.clean "sed $CLEAN_RULE"
git config filter.$FILTER_NAME.smudge "sed $SMUDGE_RULE"
git config filter.$FILTER_NAME.required true

# 检查配置是否成功
filter_check=$(git config --get filter.$FILTER_NAME.clean)
smudge_check=$(git config --get filter.$FILTER_NAME.smudge)
required_check=$(git config --get filter.$FILTER_NAME.required)
if [ -n "$filter_check" ] && [ -n "$smudge_check" ] && [ "$required_check" = "true" ]; then
  echo -e "\033[32m配置成功\033[0m"
else
  echo -e "\033[31m过滤器配置失败，请手动执行脚本内命令\033[0m"
  exit 1
fi

# ========== 【4.绑定文件到.gitattributes】 ==========
echo -e "\n\033[36m===== 绑定文件过滤规则 =====\033[0m"
for file in "${files[@]}"; do
  if ! grep -q "^${file}[[:space:]]*filter=${FILTER_NAME}$" $GIT_ATTRIBUTES 2>/dev/null; then
    echo "${file} filter=${FILTER_NAME}" >> $GIT_ATTRIBUTES
    echo -e "\033[32m绑定：$file -> $FILTER_NAME\033[0m"
  else
    echo -e "\033[33m提示：$file 已存在过滤规则\033[0m"
  fi
done

git add $GIT_ATTRIBUTES
echo -e "\033[32m修改成功：$GIT_ATTRIBUTES\033[0m"


# ========== 【5.重新追踪应用】 ==========
echo -e "\n\033[36m===== 重新追踪文件 =====\033[0m"
for file in "${files[@]}"; do
  if [ -f "$file" ]; then
    if git diff --cached --quiet "$file" >/dev/null 2>&1; then
      git rm -r --cached "$file"
    fi
    git add "$file"
  fi
done

echo -e "\033[32m完成\033[0m"
exit 0