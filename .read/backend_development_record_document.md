# 博客项目-后端开发-记录文档

**My blogweb project record documentation**

---

# 一、开发规范

## 1.1 Git 提交规范

为保障代码版本管理的规范性，提交信息需严格遵循以下标准，确保提交记录清晰可追溯：

| 提交类型 | 描述                                   | 示例                                  |
| -------- | -------------------------------------- | ------------------------------------- |
| feat     | 新增功能模块或核心功能点               | feat-user: 实现QQ第三方登录           |
| fix      | 修复功能Bug                            | fix-article: 解决文章编辑保存失败问题 |
| docs     | 文档变更（需求文档、接口文档等）       | docs-api: 补充用户模块接口说明        |
| style    | 代码格式调整（不影响功能逻辑）         | style: 修正代码缩进与分号格式         |
| refactor | 代码重构（不新增功能、不修复Bug）      | refactor-log: 优化日志模块代码结构    |
| perf     | 性能优化                               | perf-cache: 优化Redis热点数据缓存策略 |
| test     | 测试相关（新增测试用例、完善测试代码） | test-user: 补充用户登录接口测试用例   |
| build    | 构建流程或外部依赖变更                 | build: 更新GORM依赖版本               |
| ci       | 持续集成配置变更                       | ci: 调整Docker Compose部署脚本        |
| chore    | 辅助工具或构建流程变更                 | chore: 新增配置文件模板               |
| revert   | 代码回退                               | revert: 回退到feat-user提交前版本     |

## 1.2 开发时信息脱敏

采用 Git clean/smudge 过滤器机制，过滤代码仓库中的敏感配置（如数据库密码、七牛云密钥、JWT密钥等），避免敏感信息提交至远端仓库。

具体实现参考：[Git 仓库过滤敏感信息](https://juejin.cn/post/7590654280899526699)。



## 1.3 目录结构规范

严格遵循仓库标准布局，核心目录及职责如下：

| 核心目录    | 子文件/目录                                | 核心职责                                             |
| ----------- | ------------------------------------------ | ---------------------------------------------------- |
| core/       | init_cfg.go、init_logrus.go、init_db.go 等 | 核心组件初始化（配置、日志、数据库、Redis等）        |
| flags/      | enter.go                                   | 命令行参数解析，实现多环境配置加载                   |
| conf/       | conf_db.go、conf_jwt.go、conf_qiniu.go 等  | 配置结构定义，与配置文件字段映射                     |
| global/     | enter.go                                   | 全局变量存储（配置、数据库客户端等），供全项目访问   |
| api/        | user_api/、article_api/ 等                 | 接口层，定义HTTP接口与请求/响应处理                  |
| service/    | user_service/、article_service/ 等         | 业务逻辑层，封装核心业务处理逻辑                     |
| models/     | user_model.go、article_model.go 等         | 数据模型层，定义数据库表结构与ORM映射                |
| middleware/ | auth_middleware.go、log_middleware.go 等   | 中间件层，处理认证、日志、跨域等横切逻辑             |
| utils/      | ip/、cache/、file/ 等                      | 工具函数层，封装通用工具（IP解析、缓存操作等）       |
| common/     | res/、constant/ 等                         | 通用组件（统一响应、常量定义）                       |
| init/       | ipbase/、deploy/                           | 初始化资源（IP数据库文件、部署所需配置）             |
| service/email_service/、service/redis_service/redis_email/ | enter.go 等 | 邮件发送与邮箱验证码存储能力，当前未再单独拆出 `store/` 目录 |
| logs/       | -                                          | 日志记录输出文件                                     |
| test/       | -                                          | 程序测试临时文件                                     |
| uploads/    | -                                          | 文件上传目录，用于存储用户上传的文件（图片、文档等） |



## 1.4 配置文件读取规范

- **采用 YAML 格式存储配置**（`settings.yaml`）：

  - 按环境拆分配置文件（如 settings.dev.yaml、settings.prod.yaml）。

- **配置读取流程**：

  1. 通过 `flags/enter.go` 的 Parse 函数解析命令行参数，指定加载的配置文件 → 
  2. core/init_cfg.go 的 ReadCfg 函数读取配置 → 
  3. 映射为 conf/ 目录下的结构体 → 
  4. 存入 global/ 全局变量。

- **VS Code 调试配置**：在 launch.json 中设置 args 参数，避免重复输入命令行，配置示例：

  1. 首先，点击 VSCode 左侧的「运行和调试」图标（Ctrl+Shift+D）

  2. 点击「创建 launch.json 文件」按钮

  3. 选择「Go」环境

  4. 编辑生成的 **launch.json** 文件，添加或修改 args 字段

     ```json
     {
       "version": "0.2.0",
       "configurations": [
         {
           "name": "Attach to Process",
           "type": "go",
           "request": "launch",
           "mode": "auto",
           "args": ["--f=settings.yaml", "--db=false", "--version=false"]
         }
       ]
     }
     ```

  5. 按 F5 调取 launch.json 的配置来运行 main.go（注意不是使用 Code Runner）




## 1.5 接口设计规范

- 采用 **RESTful** 风格设计 API，统一接口路径命名规范：

  1. **用名词，不用动词**：接口描述的是 “资源”（如用户、订单），而非 “操作”（如查询、删除），操作由 HTTP 方法（GET/POST/PUT/DELETE）体现。
  2. **用复数形式**：表示资源集合时，优先用复数名词。
  3. **全小写，用短横线分隔**：避免大小写混合。
  4. **层级清晰，体现资源关系**：用路径层级表示资源的从属关系。
  5. **特殊场景（过滤 / 排序 / 分页）**：通过查询参数（Query String） 实现，路径仍保持资源核心。
  6. **例外场景（非资源型操作）**：极少数情况下，操作无法映射到 “资源 CRUD”（如登录、刷新令牌），可例外使用动词，但需尽量简化（如 `/login`[POST]、`/tokens/refresh`[POST]）

- 结合 HTTP 方法，对单一资源 / 资源集合的操作命名如下：

  | HTTP 方法 |   操作   |   接口示例   |                   说明                    |
  | :-------: | :------: | :----------: | :---------------------------------------: |
  |    GET    | 查询集合 |   `/users`   |               获取所有用户                |
  |    GET    | 查询单个 | `/users/123` |           获取 ID 为 123 的用户           |
  |   POST    | 创建资源 |   `/users`   |               新增一个用户                |
  |    PUT    | 全量更新 | `/users/123` |       替换 ID 为 123 的用户全部信息       |
  |   PATCH   | 部分更新 | `/users/123` | 更新 ID 为 123 的用户部分信息（如手机号） |
  |  DELETE   | 删除资源 | `/users/123` |           删除 ID 为 123 的用户           |

- 接口响应封装：

  - 在 `common/res` 中统一封装响应体，提供 Ok/Fail 系列工具函数，支持错误信息翻译（基于 validate 库），示例：

    ```go
    type Response struct {
    	Code Code   `json:"code"`
    	Data any    `json:"data"`
    	Msg  string `json:"msg"`
    }
    
    func (r Response) Json(c *gin.Context) {
      c.JSON(200, r)
    }
    
    func Ok(data any, msg string, c *gin.Context){
        Response{
    		Code: SuccessCode,
    		Data: data,
    		Msg:  msg,
    	}.Json(c)
    }
    func OkWithData(data any, c *gin.Context)
    func OkWithMsg(msg string, c *gin.Context)
    func OkWithList(list any, count int, c *gin.Context)
    func FailWithMsg(msg string, c *gin.Context)
    func FailWithData(data any, msg string, c *gin.Context)
    func FailWithCode(code Code, c *gin.Context)
    func FailWithError(err error, c *gin.Context)

- 通用查询：

  - 实现 `ListQuery[T any]` 通用列表查询方法，支持分页、排序、条件筛选。


- 接口文档：使用 ApiFox 管理接口，生成可视化文档，支持前后端协作调试。后期可生成 Swagger 文档。

# 二、环境准备

## 2.1 环境&工具选择

1. 项目名称：**myblogx**

2. 安装 Golang 环境（版本 `1.25.1`）

   ```bash
   # https://golang.org/
   go init mod "myblogx"

3. 虚拟容器环境：Docker Compose

4. 接口测试：APIFox

5. 代码编写器：VS Code | Trae | Cursor

6. 数据库：

   - MySQL（版本`5.7`) - docker镜像
   - Redis（版本`7.4.7`) - docker镜像
   - 数据库可视化管理工具：Vscode 拓展 MySQL

7. 搜索引擎：

   - Elasticsearch（版本`7.12.0`) - docker镜像

   - 搜索引擎可视化工具：Elasticvue - docker镜像

     ```bash
     docker run -p 9100:8080 --name elasticvue --restart=always -d cars10/elasticvue:1.12.0
     # http://localhost:9100/

8. 服务器数据同步：Vscode 拓展 SFTP

9. ssh 客户端：FinalShell

10. ip 数据库：ip2region

## 2.2 服务器环境

### 2.2.1 选个 “运行环境”

- 云服务器：我接下来选择用腾讯云 Ubuntu 22.04 服务器（直接用云服务器最省心）
- 备选：本地虚拟机
- 懒人版：本地电脑装 Docker（不用服务器，直接跑容器）

### 2.2.2 SSH 远程连接

云服务器默认只开部分端口，要新增访问端口许可，得去云服务器控制台的**安全组 / 防火墙**里加配置。

1. 输入以下命令，尝试密码链接：

   ```bash
   ssh 用户名@服务器IP
   ```

2. 若希望免密连接，则需要生成密钥：

   ```bash
   # 本地CMD/Git Bash执行（生成密钥）
   ssh-keygen -t rsa -b 4096
   # 一路回车（默认路径+空密码）
   ```

   - `Enter file in which to save the key`：直接回车（使用默认路径 `~/.ssh/id_rsa`）；
   - `Enter passphrase`：直接回车（设置空密码，避免每次还要输密钥密码）；
   - `Enter same passphrase again`：再次回车确认。

   执行完成后，本地会生成两个文件：

   - `~/.ssh/id_rsa`（私钥，本地保存，切勿泄露）；
   - `~/.ssh/id_rsa.pub`（公钥，需要上传到服务器）。

   ```bash
   # 把公钥传到服务器
   ssh-copy-id 用户名@服务器IP
   ```

   - Linux直接输入即可。Windows 需打开`Git bash`（安装 Git 时会自动安装）。

#### 选择 ssh 客户端

1. 可以用 **vscode** 作为 ssh 客户端：
   - 打开拓展栏，选择远程资源管理器，按下"+"，新建远程链接。

2. 也可以下载 **FinalShell**，功能稍微更多一些。

#### 美化终端

让服务器终端显示彩色文字，看信息更清楚。

1. 打开配置文件：

   ```bash
   nano ~/.bashrc
   ```

2. 找到`force_color_prompt=yes`（去掉前面的`#`注释），保存后生效：

   ```bash
   source ~/.bashrc

## 2.3 装 Docker：容器化的基础

1. 进入服务器控制台。

2. 更新服务器下载源（不知道为什么初始源出问题，改为阿里云的）

   ```bash
   sudo tee /etc/apt/sources.list > /dev/null <<EOF
   # 阿里云Ubuntu 22.04 (jammy) 镜像源
   deb http://mirrors.aliyun.com/ubuntu/ jammy main restricted universe multiverse
   deb http://mirrors.aliyun.com/ubuntu/ jammy-updates main restricted universe multiverse
   deb http://mirrors.aliyun.com/ubuntu/ jammy-backports main restricted universe multiverse
   deb http://mirrors.aliyun.com/ubuntu/ jammy-security main restricted universe multiverse
   EOF
   
   # 更新源+升级系统
   sudo apt update
   sudo apt upgrade -y
   sudo apt autoremove -y
   ```

3. 安装 docker

   ```bash
   # 卸载旧版本（可选，避免冲突）
   sudo apt remove -y docker docker-engine docker.io containerd runc
   # 安装必要的依赖工具（解决源验证、网络访问问题）
   sudo apt install -y ca-certificates curl gnupg lsb-release
   # 跳过HTTPS验证，直接下载阿里云的Docker GPG密钥（解决SSL连接重置问题）
   curl -k -fsSL http://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/docker.gpg
   # 添加阿里云Docker源（适配Ubuntu 22.04 jammy）
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/trusted.gpg.d/docker.gpg] https://mirrors.aliyun.com/docker-ce/linux/ubuntu jammy stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
   
   # 更新源索引并安装Docker
   sudo apt update -y
   sudo apt install -y docker-ce docker-ce-cli containerd.io
   
   # 启动并设置开机自启
   sudo systemctl start docker
   sudo systemctl enable docker
   
   # 验证安装
   docker --version
   sudo docker run --rm hello-world

4. 若出现docker镜像下载问题，需更新源（我用腾讯云，所以可以用内部的docker源，其他厂家服务器需另外找源）

   ```bash
   vim /etc/docker/daemon.json
   
   # 输入以下内容：
   {
     "registry-mirrors": ["https://mirror.ccs.tencentyun.com"]
   }
   
   # 更新 docker 状态
   systemctl daemon-reload
   systemctl restart docker
   
   # 再次验证
   sudo docker run --rm hello-world
   ```



## 2.4 同步代码：VSCode SFTP 插件

先理清我们的目标：通过 **服务器 / 虚拟机 + Docker** 快速搭建开发环境，使用`docker compose`实现组件化部署。

![image-20260116165300584](./assets/image-20260116165300584.png)

所以，要创建如图的目录结构，把本地`init/deploy`同步到服务器：

1. 打开 VSCode 拓展市场，搜索作者为`Natizyskunk`的 **SFTP** 插件

2. 按`Ctrl+Shift+P`选`SFTP: config`，会在 `.vscode` 里面生成一份名为 `sftp.json` 的配置文件。

3. 编辑`sftp.json`，记得删掉注释内容：

   ```json
   {
       "name": "我的服务", // 配置名称
       "host": "123.123.123.123", // 远程服务器IP
       "port": 22, // 端口，ssh默认用22
       "protocol": "sftp", // 协议，默认sftp
       "username": "root", // 远程服务器登录用户名
       "privateKeyPath": "~/.ssh/id_rsa", // 之前配置的ssh密钥
       "context": "./init/deploy",  // 本地需要同步的路径
       "remotePath": "/opt/myblogx/deploy",  // 服务器同步的存储路径
       "uploadOnSave": true,  // 保存文件时自动同步
       "downloadOnOpen": true, // 打开文件时自动同步
       "useTempFile": true, // 使用临时文件
   }
   ```

远程服务器的文件夹修改权限可能过高，要手动修改权限不然无法同步：

```bash
# 修改到对应用户可用
sudo chown -R 你的用户名:你的用户名 /opt/myblogx/myblogx_server
# 直接修改权限数字
sudo chmod -R 755 /opt/myblogx/myblogx_server
```



## 2.5 部署组件

看开头的目录图，我们把组件配置都放在`init/deploy`下，同步到远程服务器，配合`docker-compose.yml`一键启动。

### 2.5.1 启动 docker compose

配置好目录结构后，进入到服务器`docker-compose.yml`对应所在位置，输入以下命令进行启动：

（参考上方设置，我用 sftp 同步到服务器的 "`/opt/myblogx/deploy`" 路径）

```bash
cd "/opt/myblogx/deploy"

sudo rm -r "sql_master/data/" "sql_slave/data/"

sudo chmod -R 644 "./sql_master/my.cnf" "./sql_slave/my.cnf"
# sudo chown -R root:root ./

sudo docker compose up -d
```

再运行指令查看容器运行是否正常：

```bash
docker ps
```

- 如果容器状态不对，可以查看日志信息判断出错内容：

  ```bash
  docker logs -f 容器名
  ```

- 若 `port` 格式不是图中格式，只有 `3306/tcp` 则代表映射到宿主机端口失败：

  ![image-20251023134618830](./assets/image-20251023134618830.png)

  可能原因为宿主机端口被占用。可使用以下命令查看端口是否被占用：

  ```bash
  # Windows 系统
  netstat -ano | findstr "端口号"
  # Linux 系统
  netstat -tulpn | grep 端口号
  ```

  若确实被占用，则可以选择换一个映射端口；

  也可以杀死占用端口的程序（谨慎操作）：

  ```bash
  # Windows：
  taskkill /f /pid 1234
  # Linux：
  sudo kill -9 1234
  
  # 注意：将 1234 替换为实际查到的 PID
  ```

### 2.5.2 组件相关配置

#### MySQL 配置

##### 持久化配置

目录 `myblogx/init/deploy`：

- 主节点配置文件：`deploy/sql_master/my.cnf`

- 从节点配置文件：`deploy/sql_slave/my.cnf`

  > 注意，`my.cnf` 权限过松会被 mysql 放弃使用，保持一定访问权限：
  >
  > ```bash
  > chmod 644 ./sql_master/my.cnf ./sql_slave/my.cnf

- 主节点初始化配置文件：`deploy/sql_master/init.sql`

- 从节点初始化配置文件：

  - `deploy/sql_master/init.sql`
  - `deploy/sql_master/init.sh`

##### 数据一致配置

当主库中更新或删除了从库没有的数据，那么从库运行就会报错，需手动跳过错误：

```bash
# 进入从库
docker exec -it mysql-slave bash
mysql -uroot -p123456
#root 是用户名， 123456 是密码

# 中断从库复制
stop slave;

# 跳过 1 次出错语句
set global sql_slave_skip_counter=1;

# 启动从库复制
start slave;

# 查看从库状态（关键！确保Slave_IO_Running和Slave_SQL_Running都是Yes）
show slave status \G;
```

当然，这只能应急，不能让数据保持一致，后面再操作错误语句还是会出错。

要想解决数据不一致的问题，就需要

##### GTID 模式从库

如果启动 GTID 自动同步 binlog 位置，创建一个从最新位置开始复制的从库，可用以下方式

```bash
# 主库
SHOW MASTER STATUS;
```

拿到主库的 Executed_Gtid_Set

```bash
STOP SLAVE;
RESET SLAVE ALL;
RESET MASTER;

SET GLOBAL gtid_purged = '主库当前 Executed_Gtid_Set';

CHANGE MASTER TO
  MASTER_HOST='mysql-master',
  MASTER_PORT=3306,
  MASTER_USER='repl',
  MASTER_PASSWORD='你的密码',
  MASTER_AUTO_POSITION=1;

START SLAVE;
SHOW SLAVE STATUS\G
```



#### Redis 持久化配置

目录 `myblogx/init/deploy`：

- 主节点配置文件：`deploy/redis/conf/redis.conf`
- 数据存储：`deploy/redis/data`

注意，`redis.conf` 需下载官方配置文件并进行修改：

1. 进入 `conf` 文件夹，获取官方配置文件

   ```bash
   # windows 下载指令
   curl -o redis.conf https://raw.githubusercontent.com/redis/redis/7.0/redis.conf
   # linux 下载指令
   wget https://raw.githubusercontent.com/redis/redis/7.0/redis.conf -O redis.conf
   ```

2. 打开配置文件 `redis.conf` ，找到以下配置项，进行修改：

   ```bash
   # 启用AOF持久化，记录所有写操作。建议开启以提升数据安全性
   appendonly yes
   
   # 端口
   port 6379
   
   # 开启RDB持久化（默认yes，确认未被注释即可）
   save 900 1
   save 300 10
   save 60 10000
   
   # 关闭保护模式（允许外部IP连接，生产必开）
   protected-mode no
   
   # 允许后台运行（容器环境必须开启）
   daemonize no
   
   # 绑定所有IP（默认绑定127.0.0.1，修改后支持远程连接）
   bind 0.0.0.0
   
   # 设置 Redis 密码
   requirepass 123456
   ```

#### Elasticsearch：安全部署

目录 `myblogx/init/deploy/es`：

- 数据存储：`deploy/es/data`

- 插件安装（有时插入数据同步失败则：

  ```bash
  # 安装分词器（不用手动输入，已在 deploy 文件夹配好）
  cd /usr/share/elasticsearch
  bin/elasticsearch-plugin install https://get.infini.cloud/elasticsearch/analysis-ik/7.12.0

Elasticsearch 的 9200 端口是其 **REST API 通信端口**，直接暴露到公网会带来严重的安全风险，不建议开放。

部署时的正确做法是：

1. 仅内网访问，移除 docker compose 中的 ports 映射
2. 必须对外提供服务（如公网业务调用），则限制端口访问范围，仅允许业务服务器 IP 访问；同时开通 ES 内置认证。

#### Kafka 配置

目录 `myblogx/init/deploy/kafka`：

- 数据存储：`deploy/kafka/data`

注意存储目录权限

最后创建一个Topic就可以使用了

### 2.5.3 Compose 文件说明

#### 命令行方式，只运行 MySQL 镜像

```bash
cd "宿主机存放持久化数据的目录"
mkdir -p ./mysql/datadir
mkdir -p ./mysql/conf

# 映射目录并运行
docker run -itd --name mysql --restart=always -p 3307:3306 -v "宿主机存放持久化数据的目录"/conf:/etc/mysql/conf.d -v "宿主机存放持久化数据的目录"/datadir:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=root mysql:5.7
```

- `-itd`：组合参数，`-i`（交互式）保证容器的标准输入保持打开，`-t`（终端）为容器分配伪终端，`-d`（后台运行）让容器在后台以守护进程形式运行。
- `name`：指定容器的名称，方便后续通过名称管理容器。
- `p`：端口映射，端口映射，将宿主机端口(左)映射到容器端口(右)（MySQL 默认端口 3306）。
- `e`：设置环境变量
  - `MYSQL_ROOT_PASSWORD=root`：通过环境变量设置 MySQL 的 root 用户密码为 “root”。
  - `MYSQL_DATABASE=db`：通过环境变量在容器启动时自动创建名为 “db” 的数据库。
- `-v`：目录挂载（数据卷）
  - `/opt/db/mysql/conf:/etc/mysql/conf.d`：用于持久化 MySQL 的配置文件（容器内该目录的配置会同步到宿主机，避免容器删除后配置丢失）。
  - 用于持久化数据库数据（防止容器删除后数据丢失）。

##### 主从配置

1. 进主库，授权复制账号：

   ```bash
   docker exec -it mysql-master bash
   mysql -uroot -p123456
   
   # 授予复制权限：允许任何ip通过登录用户 repl 来获取复制权限，密码 123456
   GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%' identified by 'REDACTED_SHARED_PASSWORD';
   # 刷新权限
   flush privileges;
   # 查看主库binlog状态（记录File和Position，后续用）
   show master status \G;
   ```

   ![image-20260114183016077](./assets/image-20260114183016077.png)

2. 进从库，关联主库（MASTER_HOST 是主节点的 ip）：

   ```bash
   docker exec -it mysql-slave bash
   mysql -uadmin -pREDACTED_SHARED_PASSWORD
   
   # 配置主库连接信息（替换File和Position为上面查到的值）
   CHANGE MASTER TO 
     MASTER_HOST='REDACTED_SERVER_IP',
     MASTER_PORT=3306,
     MASTER_USER='repl',
     MASTER_PASSWORD='REDACTED_SHARED_PASSWORD',
     MASTER_LOG_FILE='mysql-bin.000004', # 上面的 File
     MASTER_LOG_POS=8134; # 上面的 Position
   
   # 跳过出错语句在从库的执行
   set sql_slave_skip_counter=1;
   
   # 启动从库复制
   start slave;
   
   # 查看从库状态（关键！确保Slave_IO_Running和Slave_SQL_Running都是Yes）
   show slave status \G;
   ```

   ![image-20260114183609024](./assets/image-20260114183609024.png)

3. 如果用的 MySQL 版本在 `5.7` 之后，可配置 **GTID** 模式，自动获取主库的 `File` 和 `Position`，降低配置难度和维护难度。

   ```sql
   # my.cnf 主库和从库
   gtid_mode=ON -- 开启GTID模式（全局事务标识符，自动配置）
   enforce_gtid_consistency=ON -- 强制GTID一致性（必须，确保所有事务都有GTID）
   log_slave_updates=ON -- 记录GTID到binlog（可选，级联复制必须开启）
   ```

   ```sql
   # 从库中输入修改为：
   CHANGE MASTER TO
   MASTER_HOST='mysql-master',
   MASTER_USER='repl',
   MASTER_PASSWORD='REDACTED_SHARED_PASSWORD',
   MASTER_PORT=3306,
   MASTER_AUTO_POSITION=1,
   MASTER_CONNECT_RETRY=10;
   ```

#### compose 配置，各个核心部分说明

```dockerfile
version: '3.8'   # 版本声明
services:        # 核心：定义所有容器服务
  mysql:         # 服务名（自定义，如web、mysql、redis）
    image: mysql:8.0 # 镜像来源
    restart: always  # 镜像失效后重启
    privileged: true # 容器内的root用户与宿主机root用户权限相同
    command:         # 容器启动时运行的命令
      --character-set-server=utf8mb4
    ports:       # 端口映射（宿主机：容器）
      - "3306:3306"
    volumes:     # 数据卷挂载，用于数据持久化
      - mysql-data:/var/lib/mysql
    environment: # 容器内预设环境变量
      - MYSQL_ROOT_PASSWORD=123456
    networks:    # 网络配置
      - my-network
  redis:
    ...
    image: "redis:7.4.7"
    depends_on:  # 启动依赖
      - mysql
networks:        # 定义容器连接网络
  my-network:
    driver: bridge
```



## 2.6 使用准备

### 2.6.1 使用数据库

使用 gorm 操作数据库，同时使用 dbresolver 插件进行读写分离

**gorm 常用方法**：https://www.fengfengzhidao.com/article/HXKvepIB8lppN5cbJ6HX

**gorm 官网**：https://gorm.io/zh_CN/docs/index.html

**gorm 读写分离插件**：https://github.com/go-gorm/dbresolver

```bash
go get gorm.io/plugin/dbresolver
```

**gorm/cacher 插件** ，给 gorm 启用缓存加速器和请求缓冲器：https://github.com/go-gorm/caches

```bash
go get -u github.com/go-gorm/caches/v4

# 如果想不查缓存，可用 Clauses(caches.Ignore())；这一条查询会强制跳过缓存，直接查数据库
db.Clauses(caches.Ignore()).Find(&user, "id = ?", 1)
```

涉及文件：

- **settings.yaml** ：`db / db1 / gorm` 记录配置设置
- **core/init_db.go** ：初始化 数据库 与 gorm
- **conf/conf_db.go**：数据库相关设置
  - `DSN()` 格式化数据库连接所需字段
  - `Empty()` 判断配置是否为空
- **conf/conf_gorm.go**：gorm 相关设置

### 2.6.2 使用 Redis

下载 `go-redis` 包：

```bash 
go get github.com/go-redis/redis/v8
```

可查看 https://github.com/redis/go-redis 进行操作学习

###  2.6.3 使用 Elasticsearch

下载操作包：

- **ElasticSearch 官方维护包**：

  ```bash
  # 适用于 Elasticsearch 7.x
  go get github.com/elastic/go-elasticsearch/v7

- **olivere/elastic**：曾经的主流第三方包，已停止维护，仅建议老项目兼容使用。

  ```bash
  go get github.com/olivere/elastic/v7

注意，由于 ES 比较占资源，所以一般要考虑用 **2核4G** 以上的服务器

### 2.6.4 服务器宿主机资源分配

在宿主机输入以下内容：
```bash
echo 'vm.max_map_count=262144' | sudo tee /etc/sysctl.d/99-blogx.conf
echo 'vm.overcommit_memory=1' | sudo tee -a /etc/sysctl.d/99-blogx.conf
sudo sysctl --system
```

- `vm.max_map_count=262144`
  - 提高进程可用的内存映射数量
  - 给 Elasticsearch 用
- `vm.overcommit_memory=1`
  - 允许更宽松的内存超额分配策略
  - 给 Redis 的 fork/持久化用

# 三、数据结构

**明确需求**：

- 用户管理（注册、登录、配置、社交关系）；
- 文章创作（发布、分类、配图、置顶、搜索）；
- 互动传播（点赞、收藏、评论、访问统计）；
- 辅助功能（消息通知、日志、AI 对话、运营 Banner）。

## 3.1 核心数据库表

![img](https://image.fengfengzhidao.com/pic/20240926232427.png)

| 表名                            | 核心字段                                                     | 用途说明                         |
| ------------------------------- | ------------------------------------------------------------ | -------------------------------- |
| 用户表（users）                 | id、username、password(加密)、email、avatar_url、role_id、created_at、updated_at、status | 存储用户基础信息与角色标识       |
| 角色权限表（roles）             | id、role_name、permission_ids、description                   | 管理角色与权限的关联关系         |
| 文章表（articles）              | id、title、content(Markdown)、summary、tag_ids、category_id、author_id、status(发布/草稿)、read_count、like_count、created_at、updated_at | 存储文章核心信息与统计数据       |
| 评论表（comments）              | id、article_id、user_id、content、parent_id、created_at、status(审核状态) | 存储评论信息，支持多层级回复     |
| 点赞收藏表（likes_collections） | id、user_id、target_id、target_type(文章/评论)、type(点赞/收藏)、created_at | 记录用户点赞、收藏操作           |
| 通知表（notifications）         | id、user_id、type(评论/点赞)、related_id、content、is_read、created_at | 存储用户消息通知记录             |
| 站点配置表（site_configs）      | id、config_key、config_value、description、updated_at        | 存储系统全局配置项               |
| 轮播图表（banners）             | id、image_url、link_url、sort_weight、status、created_at     | 存储首页轮播图信息               |
| 系统日志表（system_logs）       | id、log_type(登录/操作/运行)、user_id、ip、ip_location、content、created_at、level(Info/Warn/Error) | 存储系统各类日志，用于审计与排查 |

## 3.2 Redis 缓存结构

| 缓存 Key                    | Value 类型 | 过期时间             | 用途                                             |
| --------------------------- | ---------- | -------------------- | ------------------------------------------------ |
| user:session:{user_id}      | Hash       | 24h                  | 用户会话缓存，存储用户基础信息、权限标识等       |
| jwt:blacklist:{token}       | String     | Token 剩余有效期     | JWT 黑名单，实现 Token 主动注销                  |
| captcha:{code_id}           | String     | 10min                | 图形/邮箱验证码存储，用于登录、注册校验          |
| article:hot                 | List       | 1h                   | 热门文章列表（按阅读量排序），减少数据库查询压力 |
| article:detail:{article_id} | Hash       | 30min                | 缓存单篇文章详情，包含标题、内容摘要、作者信息等 |
| user:permission:{user_id}   | Set        | 24h                  | 缓存用户权限集合，快速校验接口访问权限           |
| tag:hot                     | List       | 2h                   | 热门标签列表，用于首页展示和文章筛选             |
| comment:count:{article_id}  | String     | 5min                 | 缓存单篇文章评论数，避免实时查询数据库           |
| ip:blacklist:{ip}           | String     | 1min-24h（动态调整） | 存储被拉黑的IP，防范恶意攻击                     |



# 五、功能性需求

本章以当前仓库代码实现为准，不再按设想型需求描述。已经落地的能力会给出明确接口设计表；尚未暴露 HTTP 接口但代码中已有基础设施的能力，会单独标明为“内部能力”或“预留”。

## 5.1 功能总览

### 5.1.1 角色划分

| 角色 | 枚举值 | 说明 |
| ---- | ------ | ---- |
| 管理员 | `1` | 拥有后台配置、日志管理、文章审核、标签维护等权限 |
| 普通用户 | `2` | 可登录、发文、评论、点赞、收藏、维护个人资料 |
| 访客 | `3` | 无登录态，仅可访问公开内容与公共接口 |

### 5.1.2 当前已实现模块

| 模块 | 当前状态 | 说明 |
| ---- | -------- | ---- |
| 站点配置 | 已实现 | 站点公共配置读取、后台敏感配置读取与更新 |
| 图片资源 | 已实现 | 本地上传、七牛直传 token、远程图片转存、后台图片管理 |
| 轮播图 | 已实现 | 轮播图增删改查 |
| 验证码与认证 | 已实现 | 图片验证码、邮箱验证码、邮箱注册、邮箱验证码登录、账号密码登录、QQ 登录、刷新令牌、退出登录、密码重置、邮箱绑定 |
| 用户资料 | 已实现 | 用户详情、基础信息、个人资料更新、管理员更新用户 |
| 登录日志 | 已实现 | 用户查看自己的登录日志，管理员查看全部 |
| 系统日志 | 已实现 | 日志列表、日志读取、日志删除 |
| 文章管理 | 已实现 | 文章 CRUD、审核、点赞、收藏、浏览计数、分类、标签、浏览历史 |
| 评论管理 | 已实现 | 发布评论、一级/二级评论列表、评论管理、删除、点赞 |
| 站内消息 | 已实现 | 消息列表、按条/按类型已读、按条/按类型删除、消息开关配置 |
| 定时同步 | 已实现 | Redis 计数器按天同步回 MySQL |
| 搜索与 ES 同步 | 已实现 | 已开放 `/api/search/articles`，并具备 `es_service`、`river_service`、索引/管道初始化、Binlog 同步与搜索结果二次补全能力 |
| 好友/AI/统计 | 预留 | 当前仓库暂无完整好友关系、AI 业务与独立统计路由 |

## 5.2 通用接口约定

### 5.2.1 接口基础

- 所有业务接口统一挂载在 `/api` 分组下。
- 上传后的静态资源通过 `/uploads/**` 访问。
- 统一响应结构见前文 `1.5 接口设计规范` 中 `Response` 定义。

### 5.2.2 鉴权与中间件

| 中间件 | 触发场景 | 说明 |
| ------ | -------- | ---- |
| `AuthMiddleware` | 登录接口之外的大多数写接口、部分读接口 | 从 Header 或 Query 读取 `token`，校验 JWT 和 Redis 黑名单 |
| `AdminMiddleware` | 后台管理接口 | 要求当前用户角色为管理员 |
| `CaptchaMiddleware` | 密码登录、发送邮箱验证码 | 需要请求体中带 `captcha_id`、`captcha_code` |
| `EmailVerifyMiddleware` | 邮箱注册、密码重置、邮箱绑定 | 需要请求体中带 `email_id`、`email_code`，校验通过后把邮箱写入上下文 |

### 5.2.3 通用参数与枚举

#### 分页查询公共参数

| 参数 | 类型 | 说明 |
| ---- | ---- | ---- |
| `page` | int | 页码，默认 `1` |
| `limit` | int | 每页条数，默认 `10`，最大 `100` |
| `key` | string | 关键字搜索，按模块配置做模糊匹配 |
| `order` | string | 排序字段，仅允许白名单值 |

#### 通用请求体

| 名称 | 结构 | 说明 |
| ---- | ---- | ---- |
| `IDRequest` | `{"id": 1}` | 单 ID 请求，也可通过路径参数 `:id` 传递 |
| `RemoveRequest` | `{"id_list":[1,2,3]}` | 批量删除请求 |

### 5.2.4 通用查询设计思路

项目里的列表查询不是一套 SQL 走天下，而是按复杂度拆成两种模式：

1. `ListQuery[T]`
   - 适合单表或轻量筛选场景；
   - 统一处理分页、模糊搜索、预加载、排序白名单；
   - 已用于日志、轮播图、图片、分类、标签、收藏夹、站内信等简单列表。
2. `PageIDQuery`
   - 适合文章列表、收藏夹文章列表这类“筛选复杂、排序复杂、还要回表预加载”的场景；
   - 第一阶段只查当前页主键 ID；
   - 第二阶段再按 ID 集合回表查详情；
   - 这样可以避免 `JOIN + 分页 + 排序 + Preload` 混在一条 SQL 里导致分页不稳定、重复行、排序失真等问题。

这个设计不是为了“抽象而抽象”，而是为了把简单列表和复杂列表分开治理，减少后期在文章模块上反复调 SQL 的成本。

## 5.3 日志模块

当前日志系统已经完整重构为“应用侧结构化写文件 + 采集器汇总入 ClickHouse + 后台按类型查询”的模式，不再使用旧版 `LogModel` 统一落 MySQL 的方案。

这一版重构的目标，不是单纯把“日志写出来”，而是把日志系统拆成三层能力：

1. **写入层**：应用内统一生成结构化 JSON 日志。
2. **采集层**：按日志类型切目录、切文件，由采集器异步汇总。
3. **查询层**：后台接口直接查 ClickHouse，支持按类型分页、筛选、详情查看。

这样设计的核心价值，是把“业务写日志”和“后台查日志”解耦：

- 应用侧只关心把日志写成稳定结构，不需要每次写业务都手工拼 SQL。
- 日志采集和入库走异步链路，不阻塞正常请求。
- 查询走 ClickHouse，适合做大批量日志检索，而不是继续挤占业务库。

### 5.3.1 日志体系概览

当前系统只有三类正式日志：

1. **运行日志 `runtime_logs`**
   - 记录服务运行过程中的通用事件；
   - 包括 HTTP 请求访问日志、定时任务、第三方服务调用、错误日志等。
2. **登录事件日志 `login_event_logs`**
   - 记录登录、登出、刷新令牌、注册成功/失败等认证事件；
   - 重点解决账号安全审计和登录行为追踪。
3. **操作审计日志 `action_audit_logs`**
   - 记录关键业务操作；
   - 重点解决“谁在什么时候对哪个对象做了什么操作，接口传了什么，返回了什么”。

这三类日志共用一套基础事件字段：

- `event_id`
- `ts`
- `log_kind`
- `service`
- `env`
- `host`
- `instance_id`
- `level`
- `message`
- `request_id`
- `trace_id`
- `user_id`
- `ip`
- `extra_json`

也就是说，当前日志系统不是“每种日志各写各的”，而是“公共骨架 + 类型扩展字段”的结构化设计。这样做的好处是：

- 日志格式统一，采集器配置简单；
- ClickHouse 表结构稳定，便于统一查询；
- 后面新增日志类型时，也可以复用同一套事件模型思路。

### 5.3.2 总体实现思路

当前日志链路的主线是：

1. **应用生成结构化日志**
   - 运行日志通过 `logrus` 输出；
   - 登录事件日志通过 `EmitLoginEvent` 写入；
   - 操作审计日志通过 `EmitActionAudit` 写入。
2. **按类型写入本地 JSON 行日志文件**
   - `runtime_logs`
   - `login_event_logs`
   - `action_audit_logs`
3. **Fluent Bit 监听日志目录**
   - 按类型分别采集三类日志文件；
   - 以 `JSONEachRow` 形式写入 ClickHouse。
4. **后台接口直接查 ClickHouse**
   - 管理员可按日志类型查看列表和详情；
   - 普通用户只开放自己的登录历史视图。

这套设计明确放弃了“请求时直接同步写日志库”的旧思路，原因有三点：

1. 日志是高频、追加型数据，更适合文件采集 + 分析库，而不是业务库直写。
2. 日志查询往往带筛选、时间范围、模糊搜索，ClickHouse 比 MySQL 更适合。
3. 应用侧统一写结构化 JSON 文件后，后续接 Fluent Bit、Vector、Kafka 等采集方案都更容易切换。

当前实现里，日志主链路实际是“文件 -> Fluent Bit -> ClickHouse”，并没有在日志模块里接入 Kafka 或 ES。

### 5.3.3 运行日志

#### 功能说明

- 运行日志面向“服务自身发生了什么”。
- 当前主要覆盖：
  - HTTP 请求访问日志
  - 程序运行时的 `info/warn/error` 输出
  - 定时任务、同步任务、初始化过程、第三方调用等服务级事件

#### 设计思路

运行日志沿用了 `logrus`，但不是简单打印文本，而是做了两层增强：

1. **公共字段钩子 `CommonFieldsHook`**
   - 自动补齐 `event_id`、`ts`、`log_kind`、`service`、`env`、`instance_id`、`host` 等公共字段；
   - 这样业务层就算只写一句 `Info/Error`，最终落盘时也会是完整的结构化事件。
2. **按日期切文件钩子 `FileDateHook`**
   - 运行日志会自动写入 `runtime_logs/<日期>/<app>.log`；
   - 每天自动轮转，不需要额外的日志切割脚本。

另外，HTTP 请求访问日志也不再单独写一套“请求日志表”，而是由 `LogMiddleware` 直接作为运行日志的一种事件输出：

- 每个请求进入时生成 `request_id`，同时写入响应头 `X-Request-Id`
- 请求结束后记录：
  - 方法
  - 路由
  - 状态码
  - 耗时
  - IP
  - 当前用户 ID（若能解析到 JWT）

也就是说，现在“访问日志”只是运行日志里的一个事件类型，而不是额外拆一套系统。这样能让请求级问题和服务级错误出现在同一条运行日志时间线上，排查更顺。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/logs/runtime` | GET | 管理员 | `start_at`、`end_at`、`service`、`level`、`host`、`method`、`path`、`user_id`、`page/limit/key` | 查询运行日志列表 |
| `/api/logs/runtime/:id` | GET | 管理员 | 路径参数 `id` | 查询单条运行日志详情 |

### 5.3.4 登录事件日志

#### 功能说明

- 登录事件日志面向“认证链路发生了什么”。
- 当前不仅记录登录成功/失败，还记录：
  - `login_success`
  - `login_fail`
  - `login_risk_control`
  - `register_success`
  - `register_fail`
  - `token_refresh`
  - `logout`
  - `logout_all`

#### 设计思路

登录事件日志不再区分“用户登录历史表”和“后台登录审计表”两套存储，而是统一写入 `login_event_logs`，再通过不同接口和筛选条件提供不同视图：

1. **后台管理员视图**
   - 可以按用户、IP、登录方式、事件名、成功状态等维度查询全部登录事件；
   - 重点用于安全审计和问题排查。
2. **用户个人登录历史视图**
   - `/api/users/login/log` 实际也是查 `login_event_logs`；
   - 只是额外限定为 `event_name=login_success` 且按权限约束只能看本人，管理员可切到任意用户。

这套设计的核心价值，是避免同一类认证事件在两套表里各记一份：

- 一份数据，两个读视图；
- 后台看全量认证事件；
- 用户侧看“成功登录历史”。

写入时，`EmitLoginEventFromGin` 会自动从 Gin 上下文补齐：

- `request_id`
- IP
- IP 归属地
- `UA`

业务层只需要告诉日志系统：

- 这是哪个认证事件
- 是否成功
- 用户名 / 用户 ID
- 失败原因
- 补充扩展字段

这样登录接口本身不会陷入一堆重复的“拼日志字段”代码里。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/logs/login` | GET | 管理员 | `start_at`、`end_at`、`user_id`、`ip`、`username`、`login_type`、`event_name`、`success`、`page/limit` | 查询登录事件日志列表 |
| `/api/logs/login/:id` | GET | 管理员 | 路径参数 `id` | 查询单条登录事件详情 |
| `/api/users/login/log` | GET | 登录用户 / 管理员 | `type`、`user_id`、`ip`、`start_at`、`end_at`、`page/limit` | 查询登录成功历史；`type=1` 只能查本人，`type=2` 管理员可查任意用户 |

### 5.3.5 操作审计日志

#### 功能说明

- 操作审计日志面向“关键业务动作是否可追溯”。
- 当前适用于后台修改配置、文章增删改、标签分类维护、轮播图管理、图片删除、管理员修改用户信息、全局通知等关键接口。

#### 设计思路

操作审计日志采用“路由按需采集原始数据 + 业务层补充操作语义”的方式，而不是简单把所有请求无脑全量入审计。

1. **原始请求/响应采集由中间件负责**
   - `CaptureLog(mode)` 可以按位控制是否采集：
     - 请求体
     - 响应体
     - 请求头
     - 响应头
   - 中间件会在不影响正常业务读取的前提下缓存一份原始数据。
2. **脱敏与截断在进入日志前完成**
   - 对 `password`、`token`、`authorization`、`cookie`、`email_code`、`refresh_token` 等敏感字段自动脱敏；
   - 单个 body/header 的日志体积有上限，超过会自动截断；
   - 这样既保留排查价值，又避免把敏感信息原样写进日志。
3. **业务层只补充动作语义**
   - 业务代码通过 `EmitActionAuditFromGin` 补充：
     - `action_name`
     - `target_type`
     - `target_id`
     - 是否成功
     - 请求摘要 / 响应摘要
     - 是否带上中间件采集的原始请求体、响应体、请求头、响应头
   - Gin 上下文中的用户 ID、IP、请求方法、路径、状态码、`request_id` 会自动补齐。

这套设计的重点是：把“通用采集”与“业务语义”拆开。  
中间件负责拿到原始现场，业务层负责回答“这次到底是在操作什么对象、操作成没成功”。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/logs/action` | GET | 管理员 | `start_at`、`end_at`、`user_id`、`ip`、`action_name`、`target_type`、`target_id`、`success`、`page/limit` | 查询操作审计日志列表 |
| `/api/logs/action/:id` | GET | 管理员 | 路径参数 `id` | 查询单条操作审计日志详情，详情会返回原始请求/响应快照字段 |

### 5.3.6 采集与查询链路

#### 采集链路

- 三类日志都会先落到本地 JSON 行日志文件。
- 当前目录结构分别是：
  - `runtime_logs/<日期>/<app>.log`
  - `login_event_logs/<日期>/<app>.log`
  - `action_audit_logs/<日期>/<app>.log`
- 服务启动时会主动预创建登录事件和操作审计当天日志文件，避免采集器因为“文件还不存在”持续告警。
- Fluent Bit 当前按三条 `tail` 输入分别采集三类日志，再通过 HTTP `JSONEachRow` 写入 ClickHouse。

#### 查询链路

- 后台接口直接查 ClickHouse，而不是读本地日志文件。
- 当前三张 ClickHouse 表分别是：
  - `runtime_logs`
  - `login_event_logs`
  - `action_audit_logs`
- 查询层统一支持：
  - 分页
  - 时间范围筛选
  - 类型相关字段筛选
  - 单条详情查询

其中有几个实现细节值得记住：

1. 未传 `start_at/end_at` 时，默认只查最近 7 天日志，避免后台一上来扫全表。
2. 分页上限由配置控制：
   - `query_default_limit`
   - `query_max_limit`
3. 详情查询按 `event_id` 读取，不依赖 MySQL 自增主键。

### 5.3.7 规则说明

- 当前日志模块的真实主链路是“结构化文件 + Fluent Bit + ClickHouse”，不是旧版 `LogModel`。
- 日志后台接口当前已经拆分为：
  - `/api/logs/runtime`
  - `/api/logs/login`
  - `/api/logs/action`
- 当前没有实现旧版 `/api/logs` 的统一 `log_type` 查询接口，也没有实现日志删除接口。
- `/api/users/login/log` 查询的也是 `login_event_logs`，只是固定过滤成功登录事件，并附带权限约束。
- `start_at`、`end_at` 时间格式固定为 `2006-01-02 15:04:05`。
- 运行日志中的 HTTP 请求访问日志是否开启，由 `log.request_log_enabled` 控制。
- 登录事件和操作审计日志都支持 `extra_json` 扩展字段，便于在不改主表字段的情况下补充上下文。

### 5.3.8 当前实现边界

- 当前后台查询完全依赖 ClickHouse；如果 ClickHouse 未启用，日志查询接口会直接报错，而不是自动回退去扫本地文件。
- 当前没有提供“在线浏览服务器原始日志文件”的 HTTP 接口，后台看到的是 ClickHouse 中已经采集入库的数据。
- 当前操作审计日志的原始请求/响应采集是按路由显式接入的，不是对所有接口默认全量采集。
- 当前日志系统已经统一做了敏感字段脱敏，但如果业务把敏感信息塞进自定义 `extra_json` 或非常规字段名里，仍需要业务层自己保持克制。
- 当前 `trace_id` 字段已经在结构上预留，但现阶段主要稳定使用的是 `request_id`，并没有完整接入独立分布式追踪系统。

## 5.4 站点与资源管理模块

### 5.4.1 站点配置

#### 功能说明

- 公共站点配置通过 `/api/site/site` 读取。
- 管理后台可读取并更新 `site`、`email`、`qq`、`qiniu`、`ai` 五类配置。
- 读取敏感配置时会对密钥做占位符脱敏。
- 更新 `site` 配置时，会尝试同步修改前端入口 HTML 中的标题、图标与 SEO 元信息。

#### 实现思路

> 采用 `七牛云` 对象存储：免费额度高，注册即享每月 10GB 标准存储、10GB CDN 回源流量、10 万写请求、100 万读请求，足以支撑博客 / 小型 Web 的图片、静态资源与测试场景，前期零成本启动。
>
> Go 使用文档 SDK：https://developer.qiniu.com/kodo/1238/go#5
>
> Js 使用文档 SDK：https://developer.qiniu.com/kodo/1283/javascript

- 站点配置更新不是写死 if-else 逐字段赋值，而是通过 `name -> 配置结构体` 的映射表做动态绑定。
- 对于敏感字段，读取时返回占位符；更新时如果前端仍传占位符，则沿用旧值，避免后台编辑配置时把真实密钥覆盖掉。
- `site` 配置除了更新 `settings.yaml` 外，还会用 `goquery` 修改前端 HTML 中的：
  - `<title>`
  - `favicon`
  - `keywords`
  - `description`
- 这部分设计体现的是“配置改动尽量一次生效”，不要求再手动跑前端脚本改站点标题。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/site/qq_url` | GET | 公开 | 无 | 获取 QQ OAuth 跳转地址 |
| `/api/site/:name` | GET | 公开 | `name=site` | 获取公共站点配置，当前公开场景仅实现 `site` |
| `/api/site/admin/:name` | GET | 管理员 | `name=email/qq/qiniu/ai` | 获取后台敏感配置，密钥字段脱敏后返回 |
| `/api/site/:name` | PUT | 管理员 | `name=site/email/qq/qiniu/ai` + 对应 JSON 配置体 | 更新站点配置，并回写 `settings.yaml` |

### 5.4.2 图片资源

#### 功能说明

- 当前图片系统已经不是“上传一张图片”这么简单，而是一套完整的资源生命周期系统，覆盖：
  - 上传任务创建
  - 七牛直传
  - 服务端验收
  - 正式图片入库
  - 内容审核状态同步
  - 业务引用关系维护
  - 后台删除与资源回收
- 上传链路当前以七牛云为唯一存储后端，接口层不再接收旧版 `multipart/form-data` 图片二进制。
- 图片上传完成后不会立刻默认视为“可用图片”，只有通过服务端验收并写入 `image_models` 后，才成为正式资源。
- 系统已经引入 `image_refs` 和 `image_ref_river`：
  - `image_refs` 负责记录“哪张图片被哪个业务对象引用”
  - `image_ref_river` 负责监听业务表变更，自动重建引用关系
- 当前引用跟踪已覆盖：
  - 文章封面与正文图片
  - 用户头像
  - 轮播图封面
  - 收藏夹封面

#### 实现思路

当前设计的关键思路，是把“文件传输”和“资源确权”拆开，并把“图片是否被业务使用”从上传流程中独立出来。

1. **上传任务层：先声明，再上传**
   - 前端先提交 `file_name`、`size`、`mime_type`、`hash` 这些元信息。
   - 后端先做静态校验：
     - 七牛配置是否启用且完整
     - 文件大小是否超限
     - MIME 是否属于 `image/*`
     - 文件后缀是否在白名单中
     - 哈希是否为空
   - 如果数据库里已经存在相同 `hash` 的正式图片，就直接返回 `skip_upload=true` 和现有图片信息，实现秒传。
   - 如果没有命中去重，后端会创建 Redis 上传任务，而不是先写数据库。

2. **对象存储层：客户端直传七牛**
   - 上传任务生成后，后端签发固定 `bucket + object_key` 的七牛上传凭证。
   - `object_key` 当前按 `前缀/images/日期/hash` 组织，上传凭证会同时限制：
     - 最大文件大小
     - 仅允许图片 MIME
     - `InsertOnly=1`，禁止覆盖已存在对象
   - 这样做的核心目的，是让大文件流量直接走对象存储，后端只负责签发规则和最终验收。

3. **任务状态层：Redis 短生命周期编排**
   - 上传任务不落数据库，而是暂存在 Redis。
   - Redis 中维护两类索引：
     - `task_id -> 任务详情`
     - `object_key -> task_id`
   - 这让系统同时支持两种确认方式：
     - 前端按任务 ID 查询状态
     - 七牛按对象 key 回调后端自动确认
   - 确认阶段还会对任务加 Redis 分布式锁，避免同一个任务被前端和回调并发重复处理。

4. **验收层：服务端重新核验真实对象**
   - 上传完成后，最终都会进入统一确认逻辑：
     - 七牛上传成功回调：`POST /api/images/qiniu/callback`
     - 前端手动兜底确认：`POST /api/images/upload-tasks/complete`
   - 后端不会直接信任客户端声明或回调字段，而是再次向七牛查询真实对象信息，重新校验：
     - 对象是否存在
     - 对象哈希是否与任务一致
     - 文件大小是否与声明一致
     - MIME 是否仍是图片
     - 图片格式是否在白名单中
     - 宽高信息是否可解析
   - 只有通过验收后，图片才会写入正式的 `image_models`。

5. **正式图片层：以 `image_models` 作为资源真相**
   - `image_models` 记录的是已经通过验收的正式图片，包括：
     - 上传用户
     - 存储提供方
     - bucket / object_key
     - URL
     - MIME、大小、宽高
     - 内容哈希
     - 图片状态
   - 当前 `hash` 实际保存的是七牛返回的 `etag`，因此数据库中的图片去重是以对象内容结果为准，而不是只信任前端上传前计算值。
   - 如果确认阶段发现数据库里已经有相同哈希的正式图片，系统会直接复用旧记录，并尝试删除重复上传的七牛对象，避免对象存储冗余。

6. **审核层：上传成功不等于内容可用**
   - 图片模块新增了七牛内容审核回调：`POST /api/images/qiniu/audit/callback`。
   - 审核回调会把七牛审核结论映射到系统内部图片状态，例如：
     - `pass`
     - `review`
     - `block`
     - `deleted`
     - `orphaned`
   - 如果审核回调到达时图片已经入库，就直接更新 `image_models.status`。
   - 如果审核回调先于图片入库到达，则先把审核结果暂存 Redis，等图片正式写库后再补消费，解决“审核先到、入库后到”的时序问题。

7. **引用层：上传和引用解耦，业务保存后再建关系**
   - 图片上传成功，只代表“系统里存在一张正式图片”，不代表它已经被某个业务对象使用。
   - 真正的引用关系由 `image_refs` 单独维护，字段核心是：
     - `image_id`
     - `ref_type`
     - `owner_id`
     - `field`
     - `position`
   - 也就是说，系统会额外记录“哪张图片被哪类对象、哪个字段、什么位置引用”，这样图片资源才从“文件”变成“业务资源”。

8. **引用同步层：`image_ref_river` 自动重建业务引用**
   - `image_ref_river` 是一套基于 MySQL Binlog 的引用监听服务。
   - 启动后会监听：
     - `article_models`
     - `user_models`
     - `banner_models`
     - `favorite_models`
   - 当这些表发生新增、更新、删除时，系统会自动重建或删除对应的图片引用关系。
   - 文章是当前最复杂的场景：
     - 会解析 Markdown 正文里的图片 URL
     - 会记录封面图引用
     - 正文图片引用还会写入出现顺序 `position`
   - 这种设计的核心价值，是避免在每个业务模块里都手写“图片引用同步”逻辑，把引用维护从业务保存代码里抽离出来。

9. **删除与回收层：先删对象，再清记录与引用**
   - 后台删除图片时，当前实现顺序是：
     - 先删除七牛对象
     - 再删除 `image_refs`
     - 最后物理删除 `image_models`
   - 这样能尽量保证“对象存储残留”优先被清理掉，同时避免数据库里继续保留已删除图片的引用记录。
   - 当前删除更偏后台强制管理操作，并不会先做“是否仍被业务引用”的拦截保护，因此更适合作为管理员维护接口使用。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/images/upload-tasks` | POST | 登录用户 | `file_name`、`size`、`mime_type`、`hash` | 创建上传任务；若命中去重则直接返回现有图片 |
| `/api/images/upload-tasks/complete` | POST | 登录用户 | `upload_id`、`object_key` | 手动确认上传任务完成，主要用于本地调试或回调兜底 |
| `/api/images/upload-tasks/:id` | GET | 登录用户 | 路径参数 `id` | 查询上传任务状态，供前端轮询 |
| `/api/images/qiniu/callback` | POST | 公开（由七牛回调） | 七牛回调 JSON | 七牛上传成功后的服务端确认入口 |
| `/api/images/qiniu/audit/callback` | POST | 公开（由七牛回调） | 七牛审核回调 JSON | 同步图片审核结论，更新正式图片状态 |
| `/api/images` | GET | 管理员 | `page`、`limit`、`key` | 查询图片列表，`key` 对文件名模糊搜索 |
| `/api/images` | DELETE | 管理员 | `id_list` | 批量删除图片记录 |

#### 规则说明

- 上传任务创建阶段只接收“图片元信息”，不接收文件二进制。
- `hash` 是图片内容级去重键，当前正式图片记录上的 `hash` 实际保存的是七牛返回的 `etag`。
- 上传任务属于当前登录用户，手动确认与状态查询都会校验任务归属。
- 七牛回调接口在处理前会校验回调签名，防止伪造请求。
- 只有当图片通过服务端验收后，才会写入正式的 `image_models`。
- 图片引用不会在上传阶段写入，而是在对应业务对象保存成功后，由 `image_ref_river` 重建。
- 当前只有解析出站内七牛 `object_key` 的图片 URL，才会进入引用关系维护；外链图片不会写入 `image_refs`。
- 后台删除图片时，如果任意一个七牛对象删除失败，当前删除操作会直接失败返回。

#### 当前实现边界

- 当前图片上传只接入了七牛云这一种存储提供方，`provider` 虽然做了枚举，但还没有扩展到本地存储、S3 等其他后端。
- 当前上传任务状态存在 Redis 中，不做长期持久化；它更像短生命周期的上传流程状态，而不是永久审计日志。
- 当前普通用户侧没有开放“我的图片库列表”接口，正式图片列表仍然是管理员视角。
- 当前删除接口是后台强制删除，不会因为图片仍被文章、用户、轮播图、收藏夹引用而主动拒绝操作；如果误删，业务字段中的旧 URL 仍可能成为失效地址。
- 当前 `image_ref_river` 依赖独立的 Binlog 监听配置与运行环境，部署时需要额外保证：
  - `image_ref_river.enabled` 已开启
  - MySQL Binlog 权限与 `server_id` 配置正确
  - `mysqldump` 可执行文件路径可用
- 当前模块已经不再提供旧版“后端直接收文件上传”与“远程图片转存”接口；如果后续还需要这两类能力，应以当前任务化验收模型重新接入，而不是回到旧的同步上传路径。

### 5.4.3 轮播图

#### 功能说明

- 当前轮播图模型只包含 `show`、`cover`、`href` 三个核心字段。
- 适合满足“是否展示 + 图片地址 + 跳转链接”的基础首页轮播需求。

#### 当前实现边界

- 原始设计里常见的“排序权重”字段目前尚未落地到模型与接口。
- 因此当前轮播图展示顺序主要受数据库默认查询顺序影响；如果后续需要精细控制首页轮播顺序，应补 `sort` 字段与排序规则。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/banners` | GET | 公开 | `show`、`page`、`limit`、`key` | 查询轮播图列表 |
| `/api/banners` | POST | 管理员 | `cover`、`href`、`show` | 创建轮播图 |
| `/api/banners/:id` | PUT | 管理员 | 路径参数 `id` + `cover`、`href`、`show` | 更新轮播图 |
| `/api/banners` | DELETE | 管理员 | `id_list` | 批量删除轮播图 |

## 5.5 用户与认证模块

当前认证链路已经从“单 JWT 校验”升级为一套更完整的组合方案：

- 前置校验层：图片验证码、邮箱验证码
- 风控层：密码登录失败限流、邮件发送频率限制
- 登录态层：`AccessToken + RefreshToken + UserSession`
- 失效控制层：`token_version + 会话吊销 + Redis 黑名单`

也就是说，现在的认证不再只是“签一个 token 给前端”，而是把“谁能发起登录”“登录成功后如何续期”“怎样让某个设备或全部设备立即失效”都纳入同一套设计里。

### 5.5.1 图片验证码

#### 功能说明

- 当前验证码使用 `base64Captcha` 生成。
- 是否启用由 `global.Config.Site.Login.Captcha` 控制。
- 验证码接口只负责生成 `captcha_id + base64` 图片数据。
- 密码登录与发送邮箱验证码接口会复用 `CaptchaMiddleware` 做统一校验。

#### 设计思路

- 验证码生成和验证码消费拆成两层：
  1. `GET /api/imagecaptcha` 只负责生成图片；
  2. `CaptchaMiddleware` 只负责读取请求体并校验。
- 这样做的核心目的，是把“挑战生成”和“业务接口”解耦。登录、发邮件、后续其他高风险接口都不需要重复写验证码解析逻辑。
- 当前校验调用 `global.ImageCaptchaStore.Verify(id, code, true)`，校验成功后会立即消费掉验证码，避免同一验证码被重复使用。
- 当前存储后端仍是进程内内存 `base64Captcha.DefaultMemStore`，实现简单，但天然更适合单实例；如果后续扩成多实例部署，再迁移到 Redis 这一类共享存储会更稳。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/imagecaptcha` | GET | 公开 | 无 | 返回 `captcha_id` 与 `base64` 图片数据 |

### 5.5.2 邮箱验证码、登录与会话认证

#### 功能说明

- 邮箱验证码当前支持四类用途：
  - 注册
  - 找回密码
  - 绑定邮箱
  - 邮箱验证码登录
- 登录方式当前已实现：
  - 用户名 / 邮箱 + 密码登录
  - 邮箱验证码登录
  - 邮箱注册后直接登录
  - QQ 登录
- 登录成功后，系统会同时建立服务端会话并签发双令牌：
  - `access_token`：接口鉴权使用，返回给前端
  - `refresh_token`：写入 HttpOnly Cookie，用于续期
- 当前仓库已经提供：
  - 刷新令牌接口
  - 单设备退出登录
  - 全部设备退出登录
  - 基于邮箱验证码重置密码
  - 基于旧密码修改密码
  - 绑定邮箱

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/users/email/verify` | POST | 公开 | `type`、`email`、`captcha_id`、`captcha_code` | 发送邮箱验证码，`type=1` 注册，`2` 重置密码，`3` 绑定邮箱，`4` 邮箱登录 |
| `/api/users/email/login` | POST | 公开 | `email_id`、`email_code` | 邮箱验证码登录，校验成功后签发 `access_token` 并写入 `refresh_token` Cookie |
| `/api/users/email/register` | POST | 公开 | `pwd`、`email_id`、`email_code` | 邮箱注册并直接建立登录会话 |
| `/api/users/qq` | POST | 公开 | `code` | QQ 登录，首次登录自动创建用户 |
| `/api/users/login` | POST | 公开 | `username`、`password`、`captcha_id`、`captcha_code` | 用户名或邮箱密码登录 |
| `/api/users/refresh` | POST | 公开 | 无 | 读取 `refresh_token` Cookie，刷新访问令牌并轮换刷新令牌 |
| `/api/users/password/recovery/email` | PUT | 公开 | `new_password`、`email_id`、`email_code` | 基于邮箱验证码重置密码 |
| `/api/users/password/renewal/email` | PUT | 登录用户 | `old_password`、`new_password` | 修改已绑定邮箱账户的密码 |
| `/api/users/logout` | POST | 登录用户 | 无 | 注销当前登录会话 |
| `/api/users/logout/all` | POST | 登录用户 | 无 | 注销当前用户全部登录会话 |
| `/api/users/email/bind` | PUT | 登录用户 | `email_id`、`email_code` | 绑定邮箱，邮箱信息由中间件校验后注入 |

#### 规则说明

- `POST /api/users/email/verify` 和 `POST /api/users/login` 都受 `CaptchaMiddleware` 保护；但只有当 `site.login.captcha=true` 时才真正启用。
- `POST /api/users/email/login`、`POST /api/users/email/register`、`PUT /api/users/password/recovery/email`、`PUT /api/users/email/bind` 都依赖 `EmailVerifyMiddleware`。
- `EmailVerifyMiddleware` 校验成功后，会把邮箱写入 Gin Context，业务接口不再重复解析 `email_id` / `email_code`。
- 邮箱验证码存储在 Redis 中，成功校验后会立即删除；连续输错达到上限后也会删除，避免暴力尝试。
- 发送邮箱验证码前，系统会按“邮箱 + IP + 操作类型”做频率限制，避免接口被恶意刷爆。
- 用户名密码登录支持“用户名或邮箱”两种账号字段，但只允许对“已设置密码”的用户生效。
- 密码登录前会检查账号与 IP 是否已被短时锁定；登录失败会累计账号和 IP 的失败次数，登录成功会清空对应计数。
- 所有登录方式都会校验用户状态；被禁用、封禁等不可登录状态不会允许继续签发会话。
- 所有登录方式成功后都会记录登录日志。
- `POST /api/users/refresh` 不依赖 `AuthMiddleware`，而是直接从 Cookie 中读取 `refresh_token`。
- 密码重置与密码修改都禁止“新密码等于旧密码”。
- 密码修改与密码重置成功后，会使该用户所有旧登录态失效。
- QQ 登录是否开放由 `global.Config.Site.Login.QQLogin` 控制。
- Access Token 当前支持以下请求头读取方式：
  - `Authorization: Bearer <token>`
  - `Authorization: <token>`
  - `token: <token>`
  - `Token: <token>`

#### 实现思路

当前认证链路可以拆成四层来理解：

1. **挑战校验层**
   - 图片验证码负责拦截“密码登录 / 发邮箱验证码”这类最容易被脚本滥用的入口；
   - 邮箱验证码负责把“邮箱所有权证明”抽出来，统一给注册、邮箱登录、找回密码、绑定邮箱复用。

2. **登录风控层**
   - 密码登录失败次数不直接存在数据库，而是用 Redis 记录“账号维度失败计数”和“IP 维度失败计数”；
   - 这样可以做到短时限流、低成本判断，也不会把数据库变成风控计数器。
   - 邮件发送同样走 Redis 频控，避免公开接口被高频调用。

3. **会话与令牌层**
   - 登录成功后不是只返回一个 JWT，而是：
     - 在 `user_session_models` 中落一条会话记录；
     - 把 `session_id` 和 `token_version` 写进 Access Token；
     - 把原始 `refresh_token` 写入 HttpOnly Cookie；
     - 服务端数据库只保存 `refresh_token` 的 SHA256 哈希，不保存明文。
   - 这个设计的核心价值，是把“短期接口凭证”和“长期续期凭证”分开：
     - `access_token` 适合高频接口鉴权，解析快；
     - `refresh_token` 只用于续期，且必须和数据库中的有效会话匹配。

4. **失效控制层**
   - `AuthMiddleware` 校验访问令牌时，不只验签名，而是依次校验：
     - Token 解析是否合法
     - 是否在 Redis 黑名单
     - 用户是否存在、状态是否允许登录
     - `token_version` 是否仍匹配当前用户
     - 对应 `session_id` 是否存在、未吊销、未过期
   - 这样做的好处是：
     - `token_version` 负责“用户级全量失效”，适合改密码、风控踢号；
     - `session_id` 负责“设备级精确失效”，适合退出当前设备；
     - Redis 黑名单负责“当前 access_token 立即失效”，弥补 JWT 天生无状态的问题。

围绕这四层，几个关键业务链路是这样落地的：

- **邮箱验证码链路**
  1. 发送验证码时，在 Redis 保存 `email_id -> email + code + fail_count + max_fail`；
  2. 校验时通过 Lua 脚本原子判断正确 / 错误 / 是否达最大失败次数；
  3. 校验成功就删除验证码，并把邮箱写入上下文。

- **密码登录链路**
  1. 先校验站点是否启用密码登录；
  2. 再校验 Redis 中账号 / IP 是否允许继续尝试；
  3. 查询用户名或邮箱对应的用户；
  4. 校验密码和用户状态；
  5. 创建会话、签发双令牌、写登录日志。

- **邮箱登录 / 邮箱注册 / QQ 登录链路**
  - 都复用同一套“创建服务端会话 + 返回 access token + 写 refresh cookie”的登录收口逻辑；
  - 首次自动建号时，用户名不再依赖 `MAX(id)+10000` 这类数据库推导，而是通过 Redis 自增序列 `NextAutoUsername()` 生成默认用户名，减少并发冲突和表扫描成本。

- **刷新令牌链路**
  - `POST /api/users/refresh` 只认 Cookie 中的 `refresh_token`；
  - 服务端根据其哈希找到有效会话后，生成新的 Access Token 和新的 Refresh Token；
  - 同时更新当前会话中的 `refresh_token_hash`、`expires_at`、`last_seen_at`、IP、UA、归属地信息；
  - 也就是说，刷新不是“重复发同一个长期 token”，而是“会话续期 + Refresh Token 轮换”。

- **退出登录 / 改密码链路**
  - 退出当前设备：吊销当前 `session_id`，并把当前 `access_token` 放进 Redis 黑名单；
  - 退出全部设备：吊销该用户全部会话，并把当前 `access_token` 拉黑；
  - 修改密码 / 重置密码：事务内完成“更新密码 + `token_version + 1` + 吊销全部会话”，确保旧登录态整体失效。

#### 当前实现边界

- 当前用户侧还没有开放“会话列表 / 设备管理”接口，虽然底层已经有 `user_session_models`。
- 当前图片验证码仍是进程内存存储，不适合直接拿来支撑多实例共享校验。
- 当前认证体系已支持双令牌，但还没有引入短信验证码、多因素认证、设备指纹等更强风控能力。

### 5.5.3 用户资料

#### 功能说明

- `detail` 用于返回当前登录用户完整资料与配置项。
- `base` 用于返回用户主页基础资料。
- 用户资料更新同时覆盖 `user_models` 与 `user_conf_models` 两张表。
- 用户名修改受格式校验与频率限制，当前限制为 `720` 小时内只能修改一次。
- 可更新简介、兴趣标签、头像、昵称、隐私设置、主页样式
- 管理员可手动修改用户信息，给予权限 / 改违规信息

#### 实现思路

- 用户资料更新请求里既有用户主表字段，也有用户配置表字段，因此接口内部先把请求结构拆成两个 map：
  - 一份更新 `user_models`
  - 一份更新 `user_conf_models`
- 这种设计比手写每个字段的 if 判断更利于后续扩展字段。
- 用户名修改限制不是前端提示级别，而是后端强校验：
  - 校验格式
  - 校验唯一性
  - 校验修改频率
- 这样可以避免前端被绕过时直接把非法用户名写进数据库。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/users/detail` | GET | 登录用户 | 无 | 获取当前用户完整资料与配置 |
| `/api/users/base` | GET | 登录用户 | `id` | 获取指定用户基础资料 |
| `/api/users/info` | PUT | 登录用户 | `username`、`nickname`、`avatar`、`abstract`、`like_tags`、`favorites_visibility`、`followers_visibility`、`fans_visibility`、`home_style_id` | 更新当前用户资料与配置 |
| `/api/users/admin/info` | PUT | 管理员 | `user_id`、`username`、`nickname`、`avatar`、`abstract`、`role` | 管理员更新用户基础信息与角色（避免违规信息） |

### 5.5.4 命令行用户能力

用类似以下的语法，在控制台输入信息创建新用户：

```bash
go run main.go -t user -s create
  选择角色     1 管理员   2 普通用户   3 访客
  请输入用户名   root
  请输入密码     ******
  再次输入密码   ******
  创建用户成功
```

注意，此处不能明文让密码出现在控制台

| 能力 | 命令 | 说明 |
| ---- | ---- | ---- |
| 创建 CLI 用户 | `go run . -t user -s create -f settings.yaml` | 通过命令行快速创建用户 |

## 5.6 文章管理模块

### 5.6.1 文章发布

#### 功能说明

- 发文接口负责创建文章主记录、生成 HTML 展示内容、建立分类与标签关联。
- 创建时允许两种业务状态：
  - `1=草稿`
  - `2=审核中`
- 若站点开启 `global.Config.Site.Article.SkipExamining`，则提交为“审核中”的文章会直接落为“已发布”。
- 个人博客模式下，只有管理员允许发文；社区模式下登录用户可发文。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles` | POST | 登录用户 | `title`、`abstract`、`content`、`category_id`、`tag_ids`、`cover`、`comments_toggle`、`status` | 创建文章；`status` 仅允许 `1` 草稿或 `2` 审核中 |

#### 规则说明

- `title`、`content`、`status` 为必填。
- `category_id` 若传入，必须是当前用户自己的分类。
- `tag_ids` 若传入，必须全部是“已启用标签”；重复标签会先去重。
- `abstract` 为空时，会从 Markdown 转纯文本后自动截取前 `200` 字。
- `status=2` 不一定真的入库为“审核中”，若站点配置开启免审核，会直接转成“已发布”。

#### 实现思路

- 当前正文不再采用“`content + html_content` 双存储”。
- 实际实现是：
  - 前端提交 Markdown 内容；
  - 后端通过 `utils/markdown.MdToSafe` 对内容做安全处理；
  - 处理后的安全 Markdown 直接写入 `article_models.content`。
- 也就是说，当前模型里已经没有 `html_content` 持久化字段，展示层如果需要 HTML，应基于 `content` 再做渲染，而不是依赖数据库里预存的一份 HTML 副本。
- `utils/markdown` 当前围绕 Goldmark 和安全清洗规则做内容处理，主要目标是：
  - 去掉危险 HTML / 链接等不安全内容；
  - 保留正常 Markdown 结构；
  - 为摘要提取、搜索分段、正文预处理提供统一入口。
- 文章创建与标签关系写入放在一个事务里，避免出现“文章建好了但标签关系没建好”的半成功状态。

- 标签文章数不在创建时直接回写数据库，而是记入 Redis 增量，后续由定时任务统一同步。

> #### 当前实现边界
>
> - 当前创建成功后只返回“创建文章成功”，不会直接返回新文章 ID 或整篇文章数据。
> - 原始设计里提到的“封面按 3:2 裁剪”“复制粘贴图片自动转存并在保存前强校验”目前主要依赖前端配合，后端发文接口本身没有内建裁剪与批量转存流程。
> - 当前搜索索引同步不是“完全业务双写”，也不是“完全只靠 Binlog”：
>   - 新文章建索引、删除后的索引移除，主路径仍然是 Binlog -> Canal -> `river_service`；
>   - 业务写接口目前只在部分标签 / 正文变更场景下补了 `es_service` 局部刷新。
>
> #### 前端展示
>
> 要让公式变漂亮，你必须在展示 HTML 的页面上引入一个 JavaScript 库（推荐 **KaTeX**，因为它速度最快；也可使用MathJax，与Gomarkdown匹配）。
>
> 前端（Vue SFC）对传入的 html 内容应用`:deep()`进行处理：
>
> ```css
> <style scoped>
> /* 容器样式 */
> .article-content {
> line-height: 1.6;
> color: #333;
> }
> 
> /* 穿透到动态插入的 HTML 标签 */
> .article-content :deep(h1) { font-size: 2rem; border-bottom: 1px solid #eee; }
> .article-content :deep(p) { margin: 1em 0; }
> .article-content :deep(code) { background: #f4f4f4; padding: 2px 4px; }
> 
> /* 针对你放行的数学公式类名 */
> .article-content :deep(.math) { color: #007bff; font-family: 'KaTeX_Main'; }
> </style>

### 5.6.2 文章更新

#### 功能说明

- 更新接口只允许文章作者本人修改自己的文章。
- 更新时会重新计算安全正文、摘要和标签关系。
- 如果原文章已经是“已发布”，且站点未开启免审核，那么一旦编辑会自动退回“审核中”。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/:id` | PUT | 登录用户 | 路径参数 `id` + `title`、`abstract`、`content`、`category_id`、`tag_ids`、`cover`、`comments_toggle` | 更新本人文章内容 |

#### 规则说明

- 只能更新本人文章，接口内部按 `id + author_id` 联合校验。
- 更新时仍会校验分类归属和标签启用状态，规则与发文一致。
- 已发布文章被修改时，若站点未开启免审核，会自动改为 `2=审核中`。
- 更新接口不会直接改浏览数、点赞数、收藏数、评论数等聚合计数。

#### 实现思路

- 更新前先读出旧标签集合，再和新标签集合做差集，最终把标签文章数增量写入 Redis。
- 文章字段更新与标签关联替换放在同一个事务里执行，避免正文和标签关系出现不一致。
- 如果更新了 `content`，后端会再次通过 `MdToSafe` 做安全处理；如果摘要为空，也会基于处理后的正文重新提取摘要。
- 当前更新链路同样不依赖 `html_content` 字段，而是统一维护单一的安全正文 `content`。

#### 当前实现边界

- 当前 `PUT /api/articles/:id` 请求体里没有 `status` 字段，所以它本质上是“编辑内容接口”，不是“草稿改提交审核接口”。
- 也就是说，文档层面不能把它写成“万能保存接口”；当前代码里草稿如何二次提交审核，还需要依赖现有前端流程或后续补接口能力。
- 管理员当前只有“审核”和“删除”接口，没有单独的“后台代编辑文章”接口。
- 当前业务侧 ES 局部刷新只覆盖了两类分支：
  - 改标签时调用 `UpdateESDocsTags`
  - 改正文时调用 `UpdateESDocsContent`
- 如果这次编辑只改了标题、摘要、分类、封面、评论开关，但没有同时改 `content`，搜索索引的即时刷新仍主要依赖后续 Binlog 同步。
- 如果一次同时改了正文和标签，当前代码会优先走标签刷新分支，正文检索字段的即时更新不会同步执行，仍要等待后续同步链路补齐。

### 5.6.3 文章列表

#### 功能说明

- **博客模式**：只能管理员发布文章。**社区模式**：所有人都可以发布文章
- 文章列表承担**首页列表**、**个人文章列表**、**后台全站文章管理**列表三种查询场景。
- 列表支持按作者、分类、标签、状态、标题关键字、排序方式进行组合筛选。
- 返回时会附带作者、分类、标签与计数信息。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles` | GET | 公开 / 登录 / 管理员 | `type`、`user_id`、`category_id`、`tag_id`、`status`、`page`、`limit`、`key`、`order` | 文章列表；`type=1` 查他人文章，`2` 查我的文章，`3` 管理员查全站文章 |

#### 规则说明

- `type=1` 查询他人文章时：
  - 传 `user_id` 则查具体人的文章，不填则查全部人的文章
  - 游客只能查已发布文章
  - 游客若请求 `page>1` 或 `limit>10`，会提示先登录
- `type=2` 查询我的文章时必须已登录，后端会自动把 `user_id` 改成当前用户。
- `type=3` 仅管理员可用。
- `order` 当前只开放白名单排序：
  - `view_count desc/asc`
  - `digg_count desc/asc`
  - `comment_count desc/asc`
  - `favor_count desc/asc`
- 若未显式传排序，则默认按“用户置顶优先 + 创建时间倒序”返回。

#### 实现思路

- 文章列表是当前项目里最复杂的查询之一，所以没有直接套通用 `ListQuery`，而是使用 `PageIDQuery` 做两段查询：
  1. 第一段只按筛选条件和排序规则分页拿文章 ID；
  2. 第二段按 ID 集合回表查询文章详情、作者、分类、标签。
- 这样做主要是因为列表同时面临三类复杂度：
  - 标签筛选需要 `JOIN article_tag_models`
  - 排序允许按阅读数、点赞数、收藏数等字段切换
  - 返回结果还要预加载作者、分类、标签
- 如果把这些都塞进一条 SQL 里，分页容易失真，也更难维护。
- 回表后还会把 MySQL 中的基础计数，与 Redis 中的点赞 / 收藏 / 浏览 / 评论增量叠加后再返回。

#### 当前实现边界

- 当前列表默认排序已经会叠加作者置顶和管理员置顶；返回结构里的 `user_top`、`admin_top` 都会按现有置顶关系真实返回。
- 列表接口当前返回的是安全正文 `content`，不是 `html_content`；如果后续要做真正轻量列表，可能还需要再拆更轻的 DTO。

### 5.6.4 文章详情

#### 功能说明

- 文章详情接口负责返回单篇文章的完整展示信息，包括正文内容、作者、分类、标签与计数聚合值。
- 它只负责“读内容”，不负责“加浏览量”。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/:id` | GET | 公开 / 登录 / 管理员 | 路径参数 `id` | 获取文章详情 |

#### 规则说明

- 游客只能查看已发布文章。
- 普通登录用户可以查看自己的所有文章；查看别人文章时仍只能看已发布文章。
- 管理员可查看所有文章。
- 文章详情接口不会顺手增加浏览量，浏览统计由独立接口处理。

#### 实现思路

- 详情查询会预加载：
  - 作者信息
  - 分类信息
  - 标签信息
- 响应前会把 Redis 中的点赞、收藏、浏览、评论增量与数据库基础值合并。
- 详情接口当前直接返回文章的安全正文 `content`；它并不依赖 `html_content` 这类额外持久化字段。

#### 当前实现边界

- 当前详情接口只返回聚合计数，不返回“当前登录用户是否已点赞 / 是否已收藏”这类个性化布尔态。
- 当前详情接口返回的是安全正文 `content`；如果后续要明确区分“展示态内容”和“编辑态原文”，还需要再补更清晰的字段设计或专门接口。

### 5.6.5 文章审核与删除

#### 功能说明

- 审核接口由管理员使用，负责把文章从“审核中”改成“已发布”或“已拒绝”。
- 删除分成两类：
  - 用户删除自己的文章
  - 管理员批量删除任意文章

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/:id/examine` | POST | 管理员 | 路径参数 `id` + `status`、`reason` | 审核文章；`status` 仅允许 `3` 已发布或 `4` 已拒绝 |
| `/api/articles/:id` | DELETE | 登录用户 | 路径参数 `id` | 删除本人文章 |
| `/api/articles` | DELETE | 管理员 | `id_list` | 管理员批量删除文章 |

#### 规则说明

- 审核接口只允许把文章改为：
  - `3=已发布`
  - `4=已拒绝`
- 用户删除接口只允许删除自己名下文章。
- 管理员删除接口支持批量删除，不限制文章作者。

#### 实现思路

- 审核接口当前只修改文章状态，不重写正文或标签。
- 审核通过和审核拒绝后，会异步插入系统消息，提醒文章作者。
- 删除并不是“只删文章主表”那么简单，而是依赖 `ArticleModel.BeforeDelete` 做级联清理：
  - 评论
  - 文章点赞记录
  - 收藏关系
  - 置顶关系
  - 浏览历史
  - 标签关系
- 删除标签关系时，还会同步回滚标签文章数 Redis 增量缓存，避免标签统计残留脏数据。

#### 当前实现边界

- 当前删除实现走的是直接 `Delete`，并非原文里设想的“状态型逻辑删除”。
- 管理员删除文章后，代码里目前没有额外发送“文章被删除”的系统通知；原始文档中的这一点仍属于待补能力。
- 审核拒绝接口虽然接收 `reason` 字段，但当前消息内容里还没有真正拼接这个拒绝原因，代码目前写死为空字符串。

### 5.6.6 文章互动能力

#### 功能说明

- 点赞、收藏、浏览量都采用“数据库基础值 + Redis 增量值”的读写模型。
- 文章详情与文章浏览统计分离，避免统计失败拖慢正文返回。
- 浏览量按“自然日去重”：
  - 登录用户：按 `user_id + article_id` 去重，到次日 0 点过期
  - 访客：按 `IP + User-Agent` 哈希与 `article_id` 去重，到次日 0 点过期

#### 实现思路

- 点赞与收藏都是“切换型接口”：
  - 先查关系是否存在；
  - 不存在则创建，存在则删除；
  - 然后把增量写入 Redis。
- 浏览量单独拆成接口，而不是在文章详情接口里顺手+1，主要是两个原因：
  - 文章内容接口应优先保证快和稳定；
  - 浏览统计失败不应该影响正文展示。
- 访客浏览去重使用 `md5(ip:user-agent)`，本质上是在“无需登录”的前提下取一个相对稳定但轻量的访客标识。
- 点赞、收藏成功后会异步插入站内消息，形成“业务动作成功 -> 消息异步入库”的链路。

#### 当前实现边界

- 点赞/收藏消息当前只对“同一操作人 + 同一业务对象”的重复点赞/收藏做去重；评论通知和回复通知没有同样的幂等去重逻辑。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/:id/digg` | PUT | 登录用户 | 路径参数 `id` | 点赞/取消点赞切换，目标文章必须已发布 |
| `/api/articles/view` | POST | 公开 / 登录 | `article_id` | 增加文章浏览量并维护访问历史 |

### 5.6.7 收藏夹与文章收藏

#### 功能说明

- 收藏夹属于用户个人资源，支持新建、编辑、删除、查询、批量移除收藏文章。
- 用户收藏文章时可指定收藏夹；如果未指定且默认收藏夹不存在，会自动创建“默认收藏夹”。
- 收藏夹公开性由 `user_conf_models.favorites_visibility` 控制。
- 当前实现中，查询别人的收藏夹或收藏夹内容时，会校验该用户收藏夹是否公开。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/favorite` | GET | 公开 / 登录 / 管理员 | `type`、`user_id`、`page`、`limit`、`key` | 收藏夹列表。`type=1` 查自己，`2` 查他人，`3` 管理员查任意用户 |
| `/api/articles/favorite/contents` | GET | 公开 / 登录 / 管理员 | `favorite_id`、`page`、`limit`、`key`、`order` | 查询某个收藏夹中的文章列表 |
| `/api/articles/favorite` | PUT | 登录用户 | `id`、`title`、`cover`、`abstract` | 新建或编辑收藏夹；`id=0` 视为新建 |
| `/api/articles/favorite` | DELETE | 登录用户 / 管理员 | `id_list` | 删除收藏夹 |
| `/api/articles/favorite` | POST | 登录用户 | `article_id`、`favor_id` | 收藏/取消收藏文章切换 |
| `/api/articles/favorite/contents` | DELETE | 登录用户 / 管理员 | `favorite_id`、`articles` | 从收藏夹中批量取消收藏文章 |

#### 规则说明

- 收藏夹内容列表只返回已发布文章。
- 收藏夹内容排序当前只开放 `created_at desc/asc`，表示按收藏时间排序。
- 批量删除收藏夹时，普通用户只能删自己的收藏夹；管理员可删任意收藏夹。
- 批量移除收藏文章时，普通用户也只能操作自己的收藏关系。

#### 设计思路

- 收藏夹和文章收藏关系拆开建模，而不是把“是否收藏”直接挂在文章表上，是为了支持：
  - 一个用户多个收藏夹
  - 一篇文章收藏到不同收藏夹
  - 后续扩展默认收藏夹、公开收藏夹、批量移除等能力
- `favor_id=0` 时自动创建或复用默认收藏夹，这样前台在“快速收藏”场景下无需先强制用户建收藏夹。
- 收藏夹文章列表同样采用 `PageIDQuery` 两段查询，因为它既要按收藏关系排序，又要回表查文章和作者信息。

#### 权限实现备注

- 收藏夹列表的 `type=2` 会检查目标用户是否开启了收藏夹可见。
- 分类列表的 `type=2` 当前没有像收藏夹那样再额外做隐私可见性控制，这反映的是当前代码现状，不是理想化设计。

### 5.6.8 浏览历史

#### 功能说明

- 登录用户访问文章时会通过 `ON CONFLICT` upsert 写入浏览历史表。
- 浏览历史查询和删除接口都要求登录。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/history` | GET | 登录用户 | `type`、`user_id`、`page`、`limit`、`key` | `type=1` 查自己的历史；`type=2` 按 `user_id` 查指定用户历史 |
| `/api/articles/history` | DELETE | 登录用户 | `id_list` | 删除浏览历史记录 |

#### 实现备注

- 当前 `type=2` 的浏览历史查询没有做管理员校验，也没有做用户隐私校验；这是当前实现现状，若作为正式产品能力，建议后续补权限控制。

#### 设计思路

- 浏览历史写入使用 `ON CONFLICT` upsert，而不是先查后改：
  - 避免高并发下重复插入；
  - 只更新 `updated_at` 即可表达“最近看过”。
- 这是典型的“写少量状态但要求幂等”的场景，适合直接在数据库层做冲突更新。

### 5.6.9 文章分类

#### 功能说明

- 分类是用户私有资源。
- 支持列表、创建、更新、删除、下拉选项。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/category` | GET | 公开 / 登录 / 管理员 | `type`、`user_id`、`page`、`limit`、`key` | 分类列表。`type=1` 查自己，`2` 查他人，`3` 管理员查全部 |
| `/api/articles/category` | POST | 登录用户 | `id`、`title` | 新建或编辑分类；`id=0` 视为新建 |
| `/api/articles/category` | DELETE | 登录用户 / 管理员 | `id_list` | 删除分类 |
| `/api/articles/category/options` | GET | 登录用户 | 无 | 返回当前用户分类下拉选项 |

### 5.6.10 文章置顶

#### 功能说明

- 当前项目支持两类置顶：
  - 作者置顶：作者把自己的已发布文章置顶到个人文章流前面
  - 管理员置顶：管理员把文章置顶到更高优先级的位置
- 置顶关系支持新增、取消和列表查询。
- 常规文章列表默认排序已经会叠加置顶关系，因此置顶不只是“单独查一个置顶列表”，也会影响常规文章列表的顺序。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/top` | GET | 公开 | `type`、`user_id` | 查询置顶文章列表；`type=1` 作者置顶，`2` 管理员置顶 |
| `/api/articles/top` | POST | 登录用户 | `article_id`、`type` | 置顶文章；`type=1` 作者置顶，`2` 管理员置顶 |
| `/api/articles/top` | DELETE | 登录用户 | `article_id`、`type` | 取消置顶；`type=1` 作者置顶，`2` 管理员置顶 |

#### 规则说明

- `GET /api/articles/top`
  - `type=1` 时必须传 `user_id`，表示查询某位作者的置顶文章；
  - `type=2` 时查询管理员置顶文章列表；
  - 当前置顶列表只返回已发布文章。
- `POST /api/articles/top`
  - `type=1` 时只能置顶自己的文章；
  - 普通用户作者置顶最多 `3` 篇；
  - 作者置顶要求目标文章必须已发布；
  - `type=2` 时只有管理员可以执行管理员置顶。
- `DELETE /api/articles/top`
  - `type=1` 时只能取消自己文章的作者置顶；
  - `type=2` 时只有管理员才能取消管理员置顶。

#### 设计思路

- 当前作者置顶和管理员置顶没有拆成两张表，而是共用 `user_top_article_models` 这一张关系表。
- 区分方式不是额外加一个“置顶类型字段”，而是看“是谁创建了这条置顶关系”：
  - 作者给自己的文章置顶，表现为 `user_id = article.author_id`
  - 管理员置顶，表现为 `user_id` 对应一个管理员账号
- 这样设计的好处是：
  - 数据结构更简单；
  - 新增 / 恢复 / 取消置顶都能复用同一套唯一关系逻辑；
  - 作者置顶列表和管理员置顶列表只需要切换查询条件，不需要维护两套几乎重复的模型。
- 为了避免“检查数量上限”和“创建置顶记录”之间发生并发竞态，作者置顶时会先锁当前用户行，再在事务里完成：
  1. 校验是否已置顶；
  2. 统计当前作者置顶数量；
  3. 创建或恢复置顶关系。
- 置顶状态变化后，业务侧会调用 `es_service.UpdateESDocsTop` 刷新对应文章的 ES 文档，缩短搜索侧看到旧置顶状态的时间窗口。

#### 当前实现边界

- 当前置顶列表接口没有单独做分页，返回的是当前查询条件下的全部置顶文章。
- 管理员置顶列表只返回“当前已发布”的文章；因此管理员即使先给未发布文章打了管理员置顶，这篇文章也要等发布后才会出现在置顶列表里。
- 当前作者置顶数量上限只约束普通作者；管理员如果按作者置顶自己的文章，不受 `3` 篇上限限制。

### 5.6.11 文章标签

#### 功能说明

- 标签是全站公共词典，由管理员维护。
- 用户只能在发文/改文时引用已启用标签。
- 标签列表中的文章数采用“数据库值 + Redis 增量值”合并返回。

#### 设计思路

- 标签做成全站词典，而不是用户私有标签，是为了统一检索口径和内容治理口径。
- 标签文章数不在每次文章改动后直接回写数据库，而是先走 Redis 增量，再由定时任务汇总刷库。
- 这样可以把文章编辑、审核、删除等写路径里的数据库压力削平。
- 标签标题发生变更时，后台会批量找出受影响的文章 ID，并调用 `UpdateESDocsTags` 刷新 ES 中的 `tags` 字段，尽量缩短搜索结果里的标签标题陈旧窗口。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/articles/tags/options` | GET | 登录用户 | 无 | 获取已启用标签的下拉选项 |
| `/api/articles/tags` | GET | 管理员 | `is_enabled`、`page`、`limit`、`key` | 标签后台列表 |
| `/api/articles/tags` | PUT | 管理员 | `id`、`title`、`sort`、`description`、`is_enabled` | 新建或编辑标签；`id=0` 视为新建 |
| `/api/articles/tags` | DELETE | 管理员 | `id_list` | 删除标签，若标签已被文章使用则禁止删除 |

## 5.7 评论模块

### 5.7.1 功能说明

- 当前评论系统只支持两级嵌套。
  - 网站评论普遍只做 2 级嵌套，核心是**性能、交互体验、用户认知、工程实现**四大因素的综合权衡，几乎所有主流产品（抖音、B 站、微博、知乎、Facebook）都采用这一方案。

- 一级评论列表和二级评论列表都支持分页。
- 评论是否需要审核由 `global.Config.Site.Comment.SkipExamining` 控制；管理员发表评论也会直接通过审核。
- 已发布评论会触发文章评论数缓存与根评论回复数缓存更新。
- 评论相关消息通知通过 `message_service` 异步入库。

### 5.7.2 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/comments` | GET | 公开 | `article_id`、`page`、`limit` | 查询文章一级评论列表 |
| `/api/comments/replies` | GET | 公开 | `article_id`、`root_id`、`page`、`limit` | 查询指定一级评论下的二级评论列表 |
| `/api/comments/man` | GET | 登录用户 / 管理员 | `type`、`article_id`、`user_id`、`status`、`page`、`limit`、`key` | 评论管理列表。`type=1` 查我文章下评论，`2` 查我发的评论，`3` 管理员查全部 |
| `/api/comments` | POST | 登录用户 | `content`、`article_id`、`reply_id` | 发表评论或回复评论 |
| `/api/comments/:id/digg` | POST | 登录用户 | 路径参数 `id` | 评论点赞/取消点赞切换 |
| `/api/comments/:id` | DELETE | 登录用户 / 管理员 | 路径参数 `id` | 删除评论 |

### 5.7.3 规则说明

- 发布评论前会校验：
  - 文章存在
  - 文章未关闭评论
  - `reply_id` 对应评论存在且已发布
- 回复二级评论时，仍挂到同一个一级评论 `root_id` 下。
- 删除评论权限：
  - 管理员可删全部
  - 评论作者可删自己的评论
  - 文章作者可删自己文章下的评论
- 删除一级评论时，会连带删除其下所有二级评论，并回滚文章评论数与评论点赞/回复缓存。

### 5.7.4 设计思路

- 评论系统明确只做两级嵌套：
  - 一级评论是文章主评论；
  - 二级评论是对一级评论或其他二级评论的回复；
  - 即使回复二级评论，也仍归属同一个 `root_id`。
- 这样设计的好处是：
  - SQL 和分页逻辑简单很多；
  - 前端展示结构清晰；
  - 不会陷入无限树形评论带来的性能和交互复杂度。
- 一级评论列表和二级评论列表都把 Redis 中的点赞/回复增量合并进返回值。
- 删除评论时，代码不只删评论本身，还会同步回滚：
  - 文章评论数缓存
  - 根评论回复数缓存
  - 评论点赞缓存
- 这是为了保证“删评论”不会把聚合计数留脏。

### 5.7.5 当前实现边界

- 原始设计里提到“一级评论自带两个点赞数最高的二级评论”目前没有在后端接口里落地。
- 当前实现更直接：
  - 一级评论列表接口只返回一级评论本身及其聚合计数；
  - 二级评论需要单独调用 `/api/comments/replies` 分页获取。
- 这种实现更简单，也更符合当前代码的分页与缓存结构。

## 5.8 定时任务与缓存同步

### 5.8.1 功能说明

- 使用 `gocron` 启动定时任务调度器，时区为 `Asia/Shanghai`。
- 当前每日 `02:00:00` 执行一次 Redis 计数器回写数据库任务。
- 同步过程采用“加锁 + 原子切桶”方案，避免同步时丢增量。

### 5.8.2 任务设计表

| 任务 | 调度 | Redis Key | 同步目标 | 说明 |
| ---- | ---- | --------- | -------- | ---- |
| 文章计数同步 | 每日 02:00 | `article_favorite`、`article_digg`、`article_view`、`article_comment` | `article_models.favor_count`<br />`article_models.digg_count`<br />`article_models.view_count`<br />`article_models.comment_count` | 同步文章收藏、点赞、浏览、评论数 |
| 评论点赞同步 | 每日 02:00 | `comment_digg` | `comment_models.digg_count` | 同步评论点赞数 |
| 评论回复数同步 | 每日 02:00 | `comment_reply` | `comment_models.reply_count` | 同步一级评论回复数 |
| 标签文章数同步 | 每日 02:00 | `tag_article_count` | `tag_models.article_count` | 同步标签下文章数量 |

### 5.8.3 同步机制说明

同步过程为：

```text
业务请求写入 active 哈希桶
        ↓
定时任务用 Lua 脚本把 active 原子改名为 syncing
        ↓
读取 syncing 中的增量并写回 MySQL
        ↓
写库成功后删除 syncing 桶
        ↓
若写库失败，则把增量回补到 active 桶
```

另外，每个同步任务都带 Redis 分布式锁，避免多实例重复刷库。

### 5.8.4 设计思路

- 定时同步的核心不是“定时扫 Redis”，而是“原子切桶 + 锁 + 失败回补”三件套一起工作：
  1. 用分布式锁保证同一时间只有一个实例在同步；
  2. 用 Lua 脚本把活跃桶原子改名成 `syncing` 桶；
  3. 写库失败时，把增量重新加回活跃桶，避免数据丢失。
- 写库时使用数据库表达式做原子更新，并通过 `CASE WHEN count + delta < 0 THEN 0` 防止数值被减成负数。
- 这个方案比“每次点赞立刻写库”更适合热点文章、热点评论场景，也比“直接读 Redis 全量清空”更安全。

## 5.9 站内消息模块

### 5.9.1 模块说明

- 站内消息“新增”不是通过对外 HTTP 接口手工创建，而是业务动作发生后由服务层自动生成。
- 站内消息类型当前已实现：
  - 评论文章通知
  - 回复评论通知
  - 文章点赞通知
  - 评论点赞通知
  - 文章收藏通知
  - 系统通知
  - 全局通知
- 普通站内消息实体存储在 `article_message_models`。
- 全局通知改为独立存储在：
  - `global_notif_models`：全局通知主表
  - `user_global_notif_models`：用户对全局通知的个人态表（已读 / 删除）
- 消息提醒开关不存消息表，而是挂在 `user_conf_models` 中。

因此这里需要分成两部分理解：

- **消息提醒配置**：控制用户是否接收某些类别的提醒，底层是用户配置。
- **站内消息记录**：控制评论 / 点赞 / 收藏 / 系统消息的列表展示、已读、删除，底层是 `article_message_models`。
- **全局消息**：控制管理员面向全体或特定注册阶段用户发布的公告通知，底层是“主表 + 用户个人态表”的双表模型。

#### 设计说明

- 消息发送本身不走对外 HTTP，而是由 `message_service` 在服务层异步写入。
- 这种设计把“消息产生”和“消息消费”分离开了：
  - 评论、点赞、收藏、审核等业务接口只关心业务本身成功；
  - 站内消息模块只关心消息展示、已读、删除。
- 对点赞和收藏通知，消息服务内部已做基础去重：
  - 文章点赞按 `action_user_id + type + article_id` 去重
  - 评论点赞按 `action_user_id + type + comment_id` 去重
  - 文章收藏按 `action_user_id + type + article_id` 去重

### 5.9.2 消息提醒配置

#### 功能说明

- 当前用户可以读取和修改自己的消息提醒开关。
- 这组配置本质上属于用户配置，不属于消息实体本身。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/sitemsg/conf` | GET | 登录用户 | 无 | 获取当前用户消息提醒配置 |
| `/api/sitemsg/conf` | PUT | 登录用户 | `digg_notice_enabled`、`comment_notice_enabled`、`favor_notice_enabled`、`private_chat_notice_enabled` | 更新消息提醒配置 |

#### 设计说明

- 当前配置字段实际落在 `models.UserConfModel`：
  - `digg_notice_enabled`
  - `comment_notice_enabled`
  - `favor_notice_enabled`
  - `private_chat_notice_enabled`
- 配置更新不是手写逐字段赋值，而是通过结构体字段映射转成更新 map，再统一 `Updates` 到用户配置表。
- 这也是为什么文档里不能把 `/api/sitemsg/conf` 和消息增删改查写成同一类接口，它们操作的并不是同一张表。

### 5.9.3 站内消息

#### 功能说明

- 查询接口按业务分类拉取当前用户的消息列表。
  - `t` 不是数据库原始枚举值，而是前端使用的“消息分组编号”。
- 额外提供一个“全部未读消息统计”接口，用于消息中心角标、导航栏红点等场景。

- 已读接口支持两种模式：
  - 按 `id` 标记单条消息已读
  - 按 `t` 批量标记某一类消息已读
- 删除接口同样支持两种模式：
  - 按 `id` 删除单条消息
  - 按 `t` 批量删除某一类消息

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/sitemsg` | GET | 登录用户 | `t`、`page`、`limit` | 查询站内消息列表，`t=1/2/3` 对应三类业务消息 |
| `/api/sitemsg/user` | GET | 登录用户 | 无 | 查询当前用户各类未读消息总数，返回评论、点赞/收藏、私信、系统消息四类计数 |
| `/api/sitemsg` | POST | 登录用户 | `id` 或 `t` | 标记消息已读；传 `id` 标单条，传 `t` 批量标记某一类未读消息 |
| `/api/sitemsg` | DELETE | 登录用户 | `id` 或 `t` | 删除消息；传 `id` 删单条，传 `t` 批量删除某一类消息 |

#### 规则说明

- 查询站内消息列表
  - `t=1` 表示评论类消息：
    - 评论文章通知
    - 回复评论通知

  - `t=2` 表示互动类消息：
    - 文章点赞通知
    - 评论点赞通知
    - 文章收藏通知

  - `t=3` 表示系统通知。

- 查询全部未读消息统计
  - `comment_msg_count`：评论文章通知 + 回复评论通知的未读总数
  - `digg_favor_msg_count`：文章点赞 + 评论点赞 + 文章收藏的未读总数
  - `system_msg_count`：普通系统消息未读数 + 可见全局消息中的未读数
  - `private_msg_count`：当前用户所有会话的未读私信总数

- 消息已读
  - `id` 和 `t` 不能同时为空。
  - 单条已读时，会校验该消息必须属于当前登录用户。
  - 批量已读时，只会处理当前类型下 `is_read=false` 的消息。

- 消息删除
  - `id` 和 `t` 不能同时为空。
  - 单条删除时，只能删除自己的消息。
  - 批量删除当前代码会删除该类型下当前用户的所有消息。


#### 实现思路

- 查询站内消息列表
  - 查询时会先把 `t` 映射成一组 `message_enum.Type`，再按 `receiver_id = 当前用户` 做分页查询。
  - 这种做法把前端展示分类和数据库底层枚举解耦，后续如果消息类型扩展，只要维护映射关系即可。

- 查询全部未读消息统计
  - 先统计 `article_message_models` 中 `receiver_id = 当前用户 and is_read = false` 的普通站内消息。
  - 然后按消息类型分桶累计到：
    - 评论类
    - 点赞/收藏类
    - 系统类
  - 最后再把“当前用户可见且自己未读的全局通知”并入 `system_msg_count`。
  - 这样前端在一个接口里就能拿到消息中心所有角标数据，不用自己分别请求普通消息和全局消息再做聚合。

- 消息已读
  - 单条已读会同步写入：
    - `is_read = true`
    - `read_at = 当前时间`
  - 批量已读也是统一更新 `is_read` 和 `read_at`，不是循环逐条保存。

- 消息删除：单条/批量删除自己某分类下的消息。

#### 当前实现边界

- 评论通知、回复通知默认允许重复生成；如果后续产品希望“同一人短时间连续评论/回复合并提醒”，需要在消息服务层补充去重或聚合策略。
- 系统消息支持“指定接收者”与“全员消息（`receiver_id=0`）”两种写法，但当前对外接口只覆盖个人消息查询，不包含后台群发管理能力。
- `GET /api/sitemsg/user` 中的 `private_msg_count` 当前来自 `chat_session_models.unread_count` 聚合求和，因此它反映的是“会话级未读总数”，不是逐条实时扫描消息表。
- `GET /api/sitemsg/user` 的 `system_msg_count` 已经把全局消息未读数一并算进去，因此它不只代表 `article_message_models` 里的普通系统消息。

### 5.9.4 全局消息

#### 功能说明

全局消息属于站内通信模块里的“公告型消息”，和前面的普通站内消息不同，它不是某个用户动作触发的一对一通知，而是管理员主动发布的面向多用户公告。

当前项目采用的是“全局通知主表 + 用户个人态表”的第二种实现方案，而不是“发布时给所有用户直接复制一条消息”：

- `global_notif_models` 只保存全局通知主内容
- `user_global_notif_models` 只保存某个用户对某条全局通知的个人状态
  - 是否已读
  - 是否已删除

这种做法更适合中后期用户量增长后的场景，避免管理员一发公告就立即为每个用户落一条实体消息。

#### 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/global_notif` | GET | 登录用户 / 管理员 | `type`、`page`、`limit`、`key` | 全局通知列表；`type=1` 用户查自己可见通知，`type=2` 管理员查全部通知 |
| `/api/global_notif/read` | POST | 登录用户 | `id_list` | 批量标记全局通知已读 |
| `/api/global_notif/user` | DELETE | 登录用户 | `id_list` | 用户侧批量删除全局通知 |
| `/api/global_notif` | POST | 管理员 | `title`、`content`、`icon`、`href`、`expire_time`、`user_visible_rule` | 创建全局通知 |
| `/api/global_notif` | DELETE | 管理员 | `id_list` | 管理员批量删除全局通知 |

#### 规则说明

- `type=1` 时，用户只能看到“当前仍可见、未过期、且自己没有删除”的全局通知。
- `type=2` 仅管理员可用，用于后台管理全局通知列表。
- `expire_time` 如果不传，默认是一周后过期；如果传入，不能早于当前时间 `24` 小时以内。
- `user_visible_rule` 当前支持三类可见范围：
  - `1=已注册用户可见`
  - `2=新注册用户可见`
  - `3=所有用户可见`
- 用户标记已读时，传入的通知必须满足“当前用户本来就可见”这个前提。
- 用户删除全局通知是“个人删除”，不会影响其他用户，也不会删除全局通知主表数据。
- 管理员删除全局通知时，才会真正删除主表记录，并通过模型钩子级联清理对应的用户个人态记录。

#### 实现思路

- 这里没有采用“创建公告时立刻给所有用户复制一条消息”的做法，而是延迟到用户查看 / 已读 / 删除时再维护自己的个人态记录。
- 用户列表查询时，会分两步处理：
  1. 先根据用户注册时间、公告创建时间、过期时间、可见规则，筛出当前用户理论上可见的公告；
  2. 再结合 `user_global_notif_models` 过滤掉用户已删除的公告，并补齐 `is_read` 状态。
- 可见规则不是简单的布尔开关，而是把“公告发布时间”和“用户注册时间”拿来比较：
  - `所有用户可见`：所有未过期用户都能看到
  - `已注册用户可见`：只给公告发布时已经注册的老用户看
  - `新注册用户可见`：只给公告发布后才注册的新用户看
- 用户已读和用户删除都不是改主表，而是在 `user_global_notif_models` 中创建或更新一条“个人态”记录：
  - 没记录：表示未读、未删
  - 有记录且 `is_read=true`：表示已读
  - 有记录且 `deleted_at != nil`：表示用户已删除
- 管理员删除全局通知时，依赖 `GlobalNotifModel.BeforeDelete` 自动清理关联的 `user_global_notif_models`，避免孤儿数据。

#### 当前实现边界

- 标记已读和用户删除目前都走 `id_list` 批量接口，没有“单条 read / delete 的专用路由”。
- 管理员列表接口和用户列表接口共用同一个 `GET /api/global_notif`，靠 `type` 区分场景；这在路由上更省，但文档和前端都要明确区分语义。
- 当前全局通知并没有并入 `/api/sitemsg?t=3` 的系统通知列表，而是独立一套路由和数据结构。



## 5.10 用户关系模块

### 5.10.1 模块说明

当前项目里的“用户关系”主要落地的是关注关系，而不是完整社交关系系统。

已实现的核心能力包括：

- 当前登录用户关注其他用户
- 当前登录用户取消关注其他用户
- 查看自己的关注列表
- 查看自己的粉丝列表
- 在目标用户开启可见性的前提下，查看他人的关注列表 / 粉丝列表
- 通过关注关系计算双方关系状态

当前代码里的关系状态分为四种：

- **陌生人**：双方都没有关注
- **已关注**：我关注了对方，但对方没有关注我
- **粉丝**：对方关注了我，但我没有关注对方
- **好友**：双方互关

### 5.10.2 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/follow/:id` | POST | 登录用户 | 路径参数 `id` | 当前登录用户关注指定用户 |
| `/api/follow/:id` | DELETE | 登录用户 | 路径参数 `id` | 当前登录用户取消关注指定用户 |
| `/api/follow` | GET | 登录用户 | `user_id`、`followed_user_id`、`page`、`limit` | 查询关注列表；默认查自己，也可在目标用户公开时查询他人 |
| `/api/fans` | GET | 登录用户 | `user_id`、`fans_user_id`、`page`、`limit` | 查询粉丝列表；默认查自己，也可在目标用户公开时查询他人 |

### 5.10.3 关注 / 取关

#### 功能说明

- 关注和取消关注是两条独立接口，不是“切换型”接口。
- 当前登录用户只能操作“自己 -> 别人”的关注关系。

#### 规则说明

- 不能关注自己。
- 不能取消关注自己。
- 重复关注会直接报错“请勿重复关注”。
- 未关注时取消关注，会直接报错“尚未关注该用户”。

#### 实现思路

- 关注接口先按 `(followed_user_id, fans_user_id)` 检查关系是否已存在，存在则拒绝重复写入。
- 取消关注接口会先查询这条关系，确认存在后再删除。
- 当前关注关系模型比较直接，就是一张 `user_follow_models` 关系表：
  - `fans_user_id` 表示“谁发起了关注”
  - `followed_user_id` 表示“被谁关注”

#### 当前实现边界

- 当前没有“关注/取关切换接口”，前端需要自己根据当前状态决定调 `POST` 还是 `DELETE`。
- 当前代码里还保留了 `TODO`，尚未实现“每日关注上限 / 取关上限”等风控限制。

### 5.10.4 关注列表

#### 功能说明

- 关注列表展示的是“某个用户关注了谁”。
- 默认查询当前登录用户自己的关注列表。
- 也支持查询别人的关注列表，但要受对方隐私开关控制。

#### 规则说明

- `GET /api/follow` 必须登录后才能访问，即使是查他人的关注列表也一样。
- 当 `user_id` 为空时，后端默认回填为当前登录用户。
- 当 `user_id` 不是当前登录用户时：
  - 会先读取目标用户的 `UserConfModel`
  - 若 `follow_visibility = false`，则返回“关注列表不公开”
- `followed_user_id` 可作为附加筛选条件，用来精确过滤某个关注对象。

#### 实现思路

- 查询主体仍然是 `user_follow_models`，条件是：
  - `fans_user_id = user_id`
  - 如有需要，再叠加 `followed_user_id`
- 返回时通过 `ExactPreloads` 只预加载被关注用户的必要字段：
  - `id`
  - `avatar`
  - `nickname`
  - `abstract`
  - `created_at`
- 这种写法比整表预加载更轻，适合关系列表这种字段固定的展示场景。

### 5.10.5 粉丝列表

#### 功能说明

- 粉丝列表展示的是“谁关注了某个用户”。
- 默认查询当前登录用户自己的粉丝列表。
- 也支持查询别人的粉丝列表，但同样要受对方隐私开关控制。

#### 规则说明

- `GET /api/fans` 同样必须登录。
- 当 `user_id` 为空时，默认查询当前登录用户的粉丝列表。
- 当查询他人粉丝列表时：
  - 会先读取目标用户的 `UserConfModel`
  - 若 `fans_visibility = false`，则返回“粉丝列表不公开”
- `fans_user_id` 可作为附加筛选条件，用来精确过滤某个粉丝用户。

#### 实现思路

- 查询主体仍然是 `user_follow_models`，条件是：
  - `followed_user_id = user_id`
  - 如有需要，再叠加 `fans_user_id`
- 返回时通过 `ExactPreloads` 只预加载粉丝用户的必要展示字段，而不是整表级联。

### 5.10.6 关系计算设计

#### 功能说明

- 关注模块内部还提供了 `CalUserRelationship` 和 `CalUserRelationshipBatch` 两个关系计算方法。
- 它们的作用不是直接暴露成独立 HTTP 接口，而是给别的用户展示接口复用，用来判断“当前我和这个用户是什么关系”。

#### 实现思路

- 关系判断不是分别查两次数据库，而是一次性把“我关注他”和“他关注我”两种方向的关系都查出来。
- 然后用位标记组合关系状态：
  - `iFollow = 1` 表示“我关注了对方”
  - `heFollow = 2` 表示“对方关注了我”
- 最后根据位组合映射成：
  - `1` -> 已关注
  - `2` -> 粉丝
  - `1|2` -> 好友
  - 无标记 -> 陌生人
- 批量版本 `CalUserRelationshipBatch` 会先把全部目标用户默认初始化为“陌生人”，再统一覆盖实际关系，避免调用方额外做缺省补全。

### 5.10.7 当前实现边界

- 当前模块实际只实现了“关注 / 粉丝”关系，没有实现文档早期设想中的“拉黑”“双向黑名单”“好友专属能力”等更完整社交功能。
- 当前没有单独的“关系状态查询接口”，关系计算能力主要还是内部复用方法。
- 关注列表和粉丝列表当前都要求登录后访问，并不是公开页面接口。
- 当前列表接口注释里仍有 `TODO`，尚未支持按用户名关键字搜索关注 / 粉丝用户。



## 5.11 用户私信

### 5.11.1 模块说明

当前项目的私信能力已经落地在 `api/chat_api`、`service/chat_service`、`router/chat_router.go` 这一整套实现里，不再属于预留模块。

私信模块当前包含以下能力：

- 会话列表查询
- 会话内消息列表查询
- WebSocket 长连接发送消息
- 批量标记消息已读
- 用户侧删除单条/多条消息
- 用户侧删除整个会话
- 在站内消息角标中汇总未读私信数量
- 管理员按用户维度查看会话和消息的审计视图

当前对外支持的消息类型只有三种：

- `1=文本`
- `2=图片`
- `7=Markdown`

### 5.11.2 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/chat/sessions` | GET | 登录用户 / 管理员 | `type`、`user_id`、`page`、`limit` | 查询会话列表；`type=1` 查自己，`type=2` 管理员按 `user_id` 查指定用户 |
| `/api/chat/sessions` | DELETE | 登录用户 | `session_id_list` | 用户侧批量删除会话 |
| `/api/chat/messages` | GET | 登录用户 / 管理员 | `type`、`user_id`、`session_id`、`page`、`limit` | 查询某个会话下的消息列表；`type=1` 查自己，`type=2` 管理员按 `user_id` 查指定用户 |
| `/api/chat/read` | POST | 登录用户 | `msg_id_list` | 批量标记当前用户收到的消息为已读 |
| `/api/chat/messages` | DELETE | 登录用户 | `msg_id_list` | 用户侧批量删除消息 |
| `/api/chat/ws` | GET | 登录用户 | WebSocket 消息体：`receiver_id`、`msg_type`、`content` | 建立聊天长连接并发送消息 |

### 5.11.3 会话模型与会话列表

#### 功能说明

- 私信不是按“用户对”只存一条会话，而是为双方各存一条会话记录。
- 会话列表默认查当前登录用户自己的会话，也支持管理员按用户查看。

#### 规则说明

- `GET /api/chat/sessions`
  - `type=1`：当前登录用户查自己的会话，后端会自动把 `user_id` 改成当前用户
  - `type=2`：管理员按指定 `user_id` 查看会话；若 `user_id=0` 会直接报错
- 会话列表默认按：
  - `is_top desc`
  - `last_msg_time desc`
  - `id desc`
  排序。
- 普通用户模式只返回当前仍可见的会话；管理员模式会走 `Unscoped`，因此能看到软删除会话，并额外返回 `deleted_at`。

#### 实现思路

- `ChatSessionModel` 的设计不是“单会话一条记录”，而是“同一个 `session_id`，双方各一条记录”：
  - 自己这边会话的 `user_id = 我`
  - 对端会话的 `user_id = 对方`
- 这样设计的核心好处是：查询自己的会话列表时可以直接按 `user_id` 走单表查询，不需要每次都做复杂的双方条件拼装。
- `session_id` 不是自增 ID，而是按双方用户 ID 排序后生成的稳定语义标识，例如 `chat:1:2`。这样无论谁先发消息，双方都会落到同一个逻辑会话里。
- 列表返回只预加载对端的必要展示字段：`id / nickname / avatar`，避免会话页做整用户对象预加载。

### 5.11.4 消息列表

#### 功能说明

- 消息列表按会话维度查询，不支持“全局消息流”式拉取。
- 普通用户看到的是“自己当前仍可见”的消息；管理员可以按用户查看更完整的历史视图。

#### 规则说明

- `GET /api/chat/messages`
  - `type=1`：当前登录用户查自己的会话消息
  - `type=2`：管理员按 `user_id + session_id` 查看指定用户视角下的消息
- 普通用户查消息前，必须先存在一条属于自己的会话记录，否则返回“会话不存在”。
- 普通用户模式下，消息列表会自动过滤两类消息：
  - 会话清空水位 `clear_before_msg_id` 之前的旧消息
  - 当前用户自己删除过的消息
- 管理员模式会保留这些消息，并可额外看到消息级 `deleted_at`。

#### 实现思路

- 消息列表的可见性不是只看 `chat_msg_models` 主表，而是叠加两层用户视角过滤：
  - **会话级过滤**：通过 `clear_before_msg_id` 表示“这条会话在某一时刻之前的消息都算已清空”
  - **消息级过滤**：通过 `chat_msg_user_state_models.deleted_at` 表示“当前用户单独删过这条消息”
- 这种做法的设计意义在于：
  - 删除整个会话时不用给每条旧消息都写一条删除状态
  - 删除单条消息时又能做到只影响当前用户视图
- 返回结构里：
  - `is_self` 通过 `sender_id == 当前视角 user_id` 计算
  - `is_read` 通过 `msg_status >= 已读` 计算

### 5.11.5 消息发送链路

#### 功能说明

- 当前私信发送不走 HTTP `POST`，而是统一通过 `GET /api/chat/ws` 升级 WebSocket 后发送消息。
- WebSocket 消息体当前格式为：
  - `receiver_id`
  - `msg_type`
  - `content`

#### 规则说明

- 当前只支持三种消息类型：
  - 文本
  - 图片
  - Markdown
- 发送前会校验接收人是否存在。
- 发送权限不是“登录即可发”，而是受双方关系和接收方配置影响：
  - 陌生人：只有对方开启 `stranger_msg_enabled` 才允许发送
  - 好友：可持续互发
  - 单向关系（仅关注 / 仅粉丝）：有限额地互发
- 发送前还会走 Redis 限流：
  - 同一发送用户 60 秒内最多 60 条
  - 同一会话 60 秒内最多 30 条
  - 陌生人自然周内最多 1 条
  - 单向关系在对方未回复前，自然周内最多 3 条

#### 实现思路

- 发送链路不是“先推送、后落库”，而是：
  1. WebSocket 收到消息
  2. 校验接收人、关系权限、限流额度
  3. 落库聊天消息
  4. 查找或恢复双方会话
  5. 更新双方会话最后一条消息摘要与未读数
  6. 再尝试给在线接收方推送
- 这里最关键的设计点有三个：

- **会话自动创建 / 恢复**
  - `ensureChatSessions` 会用 `(user_id, receiver_id)` 唯一键做 upsert
  - 如果用户之前删过会话，重新发消息时会把 `deleted_at` 清空，相当于把会话恢复出来

- **会话最后消息摘要**
  - 文本消息直接显示正文
  - 图片消息降级成 `[图片]`
  - Markdown 消息会先转纯文本再截断
  - 这样会话列表不需要每次再解析整条消息内容

- **发送额度预占**
  - `CheckAndReserveChatSend` 在真正落库前先预占 Redis 分钟级额度和周额度
  - 落库失败时 `Rollback`
  - 落库成功后 `Commit`
  - 这样能保证限流统计不会被失败请求白白占满

#### 当前实现边界

- 当前 WebSocket `CheckOrigin` 直接返回 `true`，是为了开发阶段前后端不同端口更容易联调；如果上线到严格生产环境，通常还需要再收紧。
- 当前发送成功后，正常场景只会向接收方在线连接推送，不会主动给发送方回一份标准成功回执；代码里这段 sender echo 目前还是注释状态。
- 当前代码是“先落库，再判断接收人是否在线”，所以前端收到“接收人不在线”时，更准确地说是“没推送到在线连接”，而不是“消息没有保存”。
- 图片消息在服务层支持更丰富的 JSON 结构，但当前 WebSocket 请求体只有一个 `content` 字段，所以对外实际只接了图片 URL 这一种最简输入。
- Markdown 消息当前虽然预留了结构化存储思路，但 `ToMarkdownChat` 还没有把标题/摘要序列化进 JSON，目前实际仍然是直接存原始 Markdown 文本。

### 5.11.6 已读与未读统计

#### 功能说明

- 用户可以批量标记自己收到的消息为已读。
- 私信未读总数会汇总到 `GET /api/sitemsg/user` 的 `private_msg_count` 中。

#### 规则说明

- `POST /api/chat/read` 只处理：
  - 当前用户是接收方
  - 且消息状态尚未到“已读”的消息
- 自己发出的消息、别人的消息、不存在的消息都会被自动忽略。
- 如果这次没有命中任何可读消息，接口会返回“没有可标记已读的消息”。

#### 实现思路

- 已读操作会在一个事务里同时完成两件事：
  - 批量把命中的消息更新为 `MsgStatusRead` 并写入 `read_at`
  - 按会话维度递减当前用户会话的 `unread_count`
- 递减未读数时不是逐条更新，而是先把本次每个 `session_id` 命中的条数聚合出来，再通过 `CASE WHEN` 批量回写，且保证不会减成负数。
- 标记已读成功后，还会把本次已读结果按“发送方 + 会话”维度分组，推送一个 `MsgTypeRead` 的 WebSocket 已读回执给发送方在线连接。
- `GET /api/sitemsg/user` 里的 `private_msg_count` 不是去消息表里实时数未读消息，而是直接对当前用户所有会话的 `unread_count` 做 `SUM`，这样消息中心角标的读取更轻。

### 5.11.7 删除设计

#### 功能说明

- 私信删除分两层：
  - 删除单条 / 多条消息
  - 删除整个会话

#### 接口设计表补充说明

- `DELETE /api/chat/messages`：删除消息，但只影响当前用户视图
- `DELETE /api/chat/sessions`：删除会话，但只影响当前用户视图

#### 实现思路

- **删除消息**
  - 不会真的删 `chat_msg_models` 主表记录
  - 而是向 `chat_msg_user_state_models` 写一条“当前用户删除了这条消息”的状态记录
  - 如果重复删除同一条消息，会通过 upsert 复用已有状态，保证幂等
- **删除会话**
  - 不会为这个会话下的每一条消息都写删除状态
  - 而是先查询该会话当前最大的消息 ID，把 `clear_before_msg_id` 推进到这个水位
  - 同时把会话 `unread_count` 置为 `0`
  - 最后再对当前用户这条会话记录做软删除

这种“消息级墓碑 + 会话级水位”双层设计，是当前私信模块最重要的实现思路之一：

- 单条删消息时足够精细
- 整体删会话时又不会产生大量状态写入
- 列表查询时只需要按水位和墓碑状态做过滤，就能还原当前用户视角

### 5.11.8 管理员视角

#### 功能说明

- 管理员可通过 `type=2` + `user_id` 查看指定用户的会话和消息视图。

#### 规则说明

- 管理员查看会话列表时，`user_id` 必填。
- 管理员查看消息列表时，`user_id` 同样必填。
- 管理员模式会走 `Unscoped`，因此能看到：
  - 已软删会话
  - 已被用户删除的消息
  - 对应的 `deleted_at`

#### 设计说明

- 这个能力本质上是“按用户视角复盘聊天数据”，不是一个独立的后台聊天系统。
- 它复用了现有会话 / 消息模型，只是在查询阶段放开软删过滤，并把用户态删除时间回填出来。

### 5.11.9 当前实现边界

- 当前没有 HTTP 版“发送私信”接口，发送链路完全依赖 WebSocket。
- 当前没有“撤回消息”“编辑消息”“会话置顶 / 静音设置修改”接口，虽然模型里已经预留了相关字段。
- 当前没有“按关键字搜索会话 / 消息”接口。
- 当前私信通知开关 `private_chat_notice_enabled` 已存在于用户配置里，但聊天模块本身还没有围绕这个开关做完整通知策略联动。
- 当前支持的消息类型枚举里还预留了语音、视频、文件、表情等类型，但对外发送链路还没有真正开放这些能力。



## 5.12 搜索模块

### 5.12.1 模块说明

当前项目已经开放文章搜索接口：`GET /api/search/articles`。

但这个模块并不是“直接把 ES 结果原样透传给前端”，而是一个三层读模型组合：

- Elasticsearch 负责全文检索、排序打分和高亮
- MySQL 负责补齐分类标题、作者昵称、作者头像等展示字段
- Redis 负责叠加尚未刷回数据库的浏览 / 点赞 / 收藏 / 评论增量

因此它本质上是“ES 检索 + MySQL 展示元数据 + Redis 实时计数”的混合搜索模块，而不是单纯的 ES 代理接口。

### 5.12.2 接口设计表

| 接口 | 方法 | 权限 | 主要参数 | 说明 |
| ---- | ---- | ---- | -------- | ---- |
| `/api/search/articles` | GET | 公开 / 登录 / 管理员 | `page`、`limit`、`key`、`type`、`sort`、`tag_list`、`category_id`、`user_id`、`top_search`、`status` | 搜索文章列表，支持普通搜索、猜你喜欢、作者文章、自己文章、管理员搜索 |

### 5.12.3 搜索模式与规则说明

#### 搜索类型

- `type=1` 普通搜索
  - 面向游客和普通用户；
  - 只搜索已发布文章。
- `type=2` 猜你喜欢
  - 路由本身不强制登录；
  - 如果能从请求里解析出 token，就会读取当前用户 `like_tags`，对匹配标签做额外加权；
  - 如果没登录或 token 无效，会自动退化成普通搜索，不直接报错。
- `type=3` 作者文章
  - 搜索指定作者的已发布文章；
  - 需要通过 `user_id` 指定作者。
- `type=4` 自己文章
  - 必须登录；
  - 默认搜索当前用户除“已删除”外的全部文章；
  - 如果显式传 `status`，则按指定状态精确筛选；
  - 当前显式禁止搜索 `已删除` 状态。
- `type=5` 管理员搜索
  - 必须是管理员；
  - 默认可查全部状态；
  - 如果传 `status`，则按状态精确筛选。

#### 排序类型

- `sort=1` 默认排序
- `sort=2` 最新发布
- `sort=3` 最多回复
- `sort=4` 最多点赞
- `sort=5` 最多收藏
- `sort=6` 最多浏览

这里要特别注意一个实现细节：当前排序不是“只按时间 / 只按计数字段”硬排，而是始终先按 `_score desc` 排，再把 `created_at / comment_count / digg_count / favor_count / view_count` 作为次级排序字段追加进去。也就是说：

- 相关性始终是第一排序因子；
- 时间和互动计数在当前实现里是“二级排序”，不是完全覆盖相关性。

#### 其他筛选规则

- `tag_list` 按标签标题筛选，内部会做：
  - 去空值
  - 去重
  - `TrimSpace`
- `category_id` 当前只允许用于：
  - `type=3` 作者文章
  - `type=4` 自己文章
- `type=3` 使用 `category_id` 时必须同时传 `user_id`。
- `top_search=true` 时：
  - `type=3/4` 会对 `author_top` 和 `admin_top` 都做加权；
  - 其他类型只对 `admin_top` 做加权。

### 5.12.4 查询构建与打分设计

#### 关键词匹配

- 有关键词时，搜索核心走 `multi_match`，匹配字段为：
  - `title`
  - `abstract`
  - `content_parts.content`
- 没有关键词时，不会构造空查询，而是改成 `match_all`，再交给综合评分函数做“推荐式排序”。

这意味着当前搜索模块既支持传统关键字检索，也兼顾了“无关键词内容流”的排序需求。

#### 综合评分

当前实现没有只依赖 ES 默认 `_score`，而是在外层再包了一层 `function_score`。打分主要由两部分构成：

- 新鲜度衰减：对 `created_at` 做 `gauss decay`
- 互动信号加权：对点赞、评论、收藏、浏览计数做 `log1p` 权重叠加

这和原始设计文档里“搜索不是简单模糊匹配，还要体现内容热度与时效性”的思路是一致的。当前项目最终采用的不是独立推荐引擎，而是在 ES 查询层内先把这套综合排序收敛掉。

#### 为什么“猜你喜欢”不用单独建推荐表

- 当前 `猜你喜欢` 的实现很克制，没有引入独立推荐服务；
- 它只是读取用户配置里的 `like_tags`，然后在原有搜索查询上追加一个 `should terms(tags.id)` 加权。

这样做的优点是：

- 实现简单，不需要额外的推荐离线任务；
- 和搜索共用同一套 ES 文档与接口；
- 匿名用户还能自然退化为普通搜索，不需要专门兜底接口。

### 5.12.5 搜索结果组装设计

#### 高亮与返回字段

当前搜索结果不是直接返回整篇正文，而是按“搜索列表 DTO”裁剪后的结构返回。

- 标题支持高亮
- 摘要支持高亮
- 正文命中时，优先返回正文片段高亮
- `Content` 字段返回的是命中摘要 / 预览片段，不是完整正文

另外还有一个很细的取舍：

- 当 `key` 为空时，`_source` 不会请求 `content_head` 和 `content_parts`
- 高亮字段里也不会包含正文相关字段

这样做是为了避免“无关键词推荐流”把大段正文和分段结构一并返回，造成 ES 传输体积和接口响应体无意义膨胀。

#### 为什么还要回查 MySQL

ES 文档里当前只存检索和过滤强相关字段，例如：

- 标题、摘要、正文分段
- 分类 ID、作者 ID
- 标签
- 计数字段
- 状态
- 置顶标记

但前台列表展示还需要：

- `category_title`
- `user_nickname`
- `user_avatar`

这些字段当前没有直接冗余进 ES，而是搜索结果拿到文章 ID 后，再批量回查 MySQL 统一补齐。这样做的好处是：

- 作者改昵称、改头像时，不需要额外刷新所有文章 ES 文档；
- 分类改名也不会强迫 ES 全量回补；
- ES 只存“检索必要字段”，索引更稳定。

#### 为什么还要叠加 Redis

ES 里的文章计数来自索引文档，而项目里的点赞 / 收藏 / 浏览 / 评论计数当前采用“数据库基础值 + Redis 增量值”模型。

所以搜索结果在返回前，会再批量读取 Redis，把这 4 类增量叠加回去，避免出现：

- 刚点赞了，搜索结果里还看不到最新数字
- 定时刷库前，搜索列表和文章详情计数口径不一致

#### 正文分段为什么只返回部分字段

当前 `Part` 字段并不会返回完整正文，而是只提取分段结构中的：

- `level`
- `title`
- `path`

这说明它更偏向“命中位置 / 内容结构提示”，不是拿来替代文章详情正文。

### 5.12.6 ES 文档与同步链路

#### ES 文档结构

当前文章搜索索引围绕 `models.ArticleModel` 定义，主要检索字段包括：

- `title`
- `abstract`
- `content_head`
- `content_parts`
- `category_id`
- `author_id`
- `tags`
- `status`
- `view_count / digg_count / comment_count / favor_count`
- `admin_top / author_top`

这里有两个很重要的实现点：

- `content_head` 不是数据库真实列，而是 ES 冗余字段，用于保存正文纯文本前缀；
- `content_parts` 也不是原始 Markdown，而是后端预处理后的正文分段结构。

这两块都由 Go 服务端在写 ES 文档前通过 `markdown` 工具预生成，而不是依赖 ES Pipeline 在线拆分。

#### Pipeline 的真实作用

虽然当前索引初始化里仍然保留了 `article_pipeline`，但它现在实际上是一个 no-op pipeline。原因是：

- 正文摘要抽取已经在服务端完成；
- 正文分段也已经在服务端完成；
- ES Pipeline 当前更像是一个“索引契约占位”，而不是承担真正的数据清洗逻辑。

#### 当前同步方案

结合原始设计文档里的几种方案对比，当前项目并没有选“纯同步双写”或“MQ 异步双写”，而是形成了一个更贴近现状的组合方案：

- 主链路：Binlog -> Canal -> `river_service`
- 补充链路：业务写后调用 `es_service` 做局部字段刷新

这套组合方案的实际含义是：

- 建索引、删索引、全量兜底同步，主要依赖 Binlog 订阅链路；
- 某些高频改动场景，业务层会直接补一段 ES 局部更新，缩短搜索结果变旧的时间窗口。

#### 当前已经接入的局部刷新

- 文章标签变更后，调用 `UpdateESDocsTags`
- 文章正文变更后，调用 `UpdateESDocsContent`
- 标签标题改名后，会找到受影响文章，再批量调用 `UpdateESDocsTags`

这说明当前仓库的搜索同步，已经不再是文档旧版本里描述的“完全只靠 Binlog”。

### 5.12.7 能力表

| 能力 | 当前状态 | 说明 |
| ---- | -------- | ---- |
| ES 客户端初始化 | 已实现 | 支持连接 Elasticsearch 7.x |
| 索引 / Mapping / Pipeline 初始化 | 已实现 | 支持文章索引和 Pipeline 创建、删除；当前 Pipeline 为 no-op |
| MySQL Binlog 同步 ES | 已实现 | 基于 Canal 订阅方式同步 |
| 对外文章搜索 HTTP API | 已实现 | 已开放 `GET /api/search/articles` |
| 搜索高亮 | 已实现 | 支持标题、摘要、正文片段高亮 |
| 个性化猜你喜欢 | 已实现基础版 | 基于用户 `like_tags` 做标签加权，不是独立推荐系统 |
| 搜索结果实时计数补偿 | 已实现 | 返回前叠加 Redis 增量值 |

### 5.12.8 当前实现边界

- 当前搜索只覆盖文章，没有评论搜索、用户搜索、私信搜索等接口。
- 搜索路由虽然是公开接口，但 `type=4` 和 `type=5` 仍依赖 token 权限；`type=2` 未登录时会自动退化为普通搜索。
- 当前业务侧 ES 局部刷新并不完整：
  - 新文章创建后没有直接调用 `SyncESDocs`
  - 删除文章后也没有直接调用 `DeleteDocument`
  - 但文章置顶 / 取消置顶已经会调用 `UpdateESDocsTop` 做局部刷新
- 因此从系统真实行为上看，搜索一致性的主兜底仍然是 Binlog 同步链路。



## 5.13 命令模块

### 5.13.1 模块说明

`flags/` 不是面向前台或后台页面的业务模块，而是整个后端程序的命令行入口分流层。它的职责很明确：先解析启动参数，再决定当前进程是进入“正常启动 Web 服务”模式，还是进入“执行一次性运维动作后退出”模式。

当前项目里，这个模块主要承担 4 类工作：

- 解析配置文件路径，决定本次启动读取哪份 `settings.yaml`
- 执行数据库表结构迁移
- 初始化或删除 Elasticsearch 索引 / Pipeline
- 在终端交互式创建用户

它在启动链路中的位置很靠前，但又不是最前。当前实现是：`main.go` 先完成参数解析、配置读取以及核心依赖初始化，再调用 `flags.Run()` 做命令分流；如果命中了命令行子任务，就执行完成后 `os.Exit(0)`，否则继续进入 `cron` 和 `router`。

### 5.13.2 命令行能力表

| 参数 / 组合 | 当前状态 | 入口 | 说明 |
| ----------- | -------- | ---- | ---- |
| `-f settings.yaml` | 已实现 | `flags.Parse()` | 指定配置文件路径，默认值为 `settings.yaml` |
| `-db` | 已实现 | `flags.FlagDB()` | 执行 GORM `AutoMigrate`，完成后退出进程 |
| `-es -s init` | 已实现 | `flags.FlagESIndex()` | 交互式初始化 / 删除文章索引与 Pipeline，完成后退出进程 |
| `-es -s article-sync` | 已实现 | `flags.FlagESArticleSync()` | 按批次全量同步文章数据到 ES，完成后退出进程 |
| `-t user -s create` | 已实现 | `FlagUser.Create()` | 交互式创建命令行用户，完成后退出进程 |
| `-version` | 仅解析未生效 | `flags.Parse()` | 参数已定义，但当前 `flags.Run()` 没有输出版本信息的实现 |

### 5.13.3 启动分流设计

#### 启动顺序

当前主程序的启动顺序可以概括为：

1. `flags.Parse()` 解析命令行参数；
2. 将配置文件路径保存到 `global.Flags.File`；
3. 依次初始化 `global.Config`、`global.Logger`、`global.Redis`、`global.KafkaMysqlClient`、`global.DB`、`global.ESClient`；
4. 调用 `flags.Run()` 判断是否执行一次性命令；
5. 如果没有命中命令任务，再继续启动 MySQL -> ES 同步、定时任务和 HTTP 服务。

#### 设计思路

- 这个设计的核心思路不是把 `flags` 做成复杂的 CLI 框架，而是做一个“进程入口分流器”。
- `-db`、`-es`、`-t user -s create` 这些动作都属于“启动前 / 运维期工具能力”，它们和常规 HTTP 路由是并列关系，不应该混进业务接口层。
- 当前把 `flags.Run()` 放在基础设施初始化之后，是因为这些命令本身就依赖数据库、日志器、ES 客户端等运行时资源。例如：
  - `FlagDB` 需要真实的 `*gorm.DB`
  - `FlagESIndex` 需要 ES 连接和日志输出
  - `FlagUser.Create` 需要数据库查询用户名是否重复，并写入用户表

#### 当前取舍

- 这种写法实现简单，业务上足够直接；
- 但代价也很明显：即使只是执行 `-db` 迁移，当前进程也会先初始化 Redis、Kafka、ES 等组件，启动成本偏高。

### 5.13.4 数据库迁移命令

#### 功能说明

`go run . -db -f settings.yaml` 会调用 `flags.FlagDB(db)`，使用 GORM 的 `AutoMigrate` 对当前项目的核心表执行结构迁移。

#### 实现思路

- 迁移列表是显式写死在 `flags/flag_db.go` 中的，而不是自动扫描模型目录。
- 这样做的好处是迁移范围可控，新增模型时必须人工确认是否进入迁移清单，避免“只是定义了模型但并不希望落表”的情况。
- 当前迁移对象已经覆盖用户、文章、评论、日志、站内消息、全局消息、关注关系、私信会话 / 消息等主要表。

#### 当前实现边界

- 当前仍然是 `AutoMigrate` 方案，不是版本化 migration 工具。
- 它适合开发期和中小型项目快速迭代，但不擅长复杂字段变更、历史版本回滚和精细化 DDL 审核。

### 5.13.5 ES 初始化命令

#### 功能说明

`go run . -es -s init -f settings.yaml` 会进入交互式命令，分别处理两类 ES 资源：

- 文章索引（index）
- 文章处理管道（pipeline）

#### 实现思路

- 这里没有把索引名、Pipeline 名字写死在命令里，而是从 `models.ArticleModel` 上读取：
  - `Index()`
  - `PipelineName()`
  - `Mapping()`
  - `Pipeline()`
- 这样 ES 结构定义和文章搜索模型绑在一起，索引初始化逻辑不会和模型元数据脱节。
- 执行时分两段交互：先决定索引创建 / 删除，再决定 Pipeline 创建 / 删除，避免一次输入同时处理多个资源时可读性过差。

#### 当前实现边界

- 当前只覆盖文章检索相关索引，没有扩展到评论、用户或聊天消息等其他搜索场景。
- 当前是交互式 `fmt.Scanln`，更适合人工运维，不适合 CI/CD 中的非交互自动化流程。
- 如果要做文章全量补索引，当前还需要显式走 `go run . -es -s article-sync -f settings.yaml`。

### 5.13.6 命令行用户创建

#### 功能说明

`go run . -t user -s create -f settings.yaml` 可在终端直接创建用户，主要面向初始化管理员或手工补录账号的场景。

#### 实现思路

- 先让操作者按枚举值选择角色，并校验角色范围是否合法；
- 再读取用户名，先查库判断是否已存在；
- 密码采用 `terminal.ReadPassword` 读取，避免在终端明文回显；
- 最后通过 `pwd.GenerateFromPassword` 完成密码加密，写入 `user_models`。

创建出的用户还有两个固定设计：

- `Nickname` 默认写为“命令用户”
- `RegisterSource` 固定为 `RegisterTerminalSourceType`

这意味着命令行建号在数据侧是可以和普通前台注册区分开的，后续做审计时能追踪来源。

### 5.13.7 当前实现边界

- `-version` 当前只是参数占位，并没有真正输出 `global.Version`。
- `op.Type` 目前只实现了 `user`；但 `op.Sub` 不再只有 `create`，ES 分支已经支持 `init` 和 `article-sync`，整体子命令体系仍然偏轻量。
- 当前命令执行失败大多是直接打印错误并退出，没有形成统一的 CLI 错误码和帮助信息体系。
- 整个 `flags` 模块本质上是“工程运维入口”，不是对外业务能力，因此文档里不应把它描述成 HTTP 功能模块。


## 5.14 Global 模块

### 5.14.1 模块说明

`global/` 是项目的全局运行时依赖容器。它本身不承载业务逻辑，而是为 `api`、`middleware`、`service`、`router`、`core` 等各层提供共享的运行时对象。

当前项目把下面这些内容统一放到了 `global` 包中：

- 启动参数记录
- 运行时配置对象
- 日志器
- 数据库连接
- Redis 客户端
- Kafka MySQL 同步客户端
- Elasticsearch 客户端
- 图形验证码存储器

可以把它理解成当前项目里最轻量的一层“全局依赖注册表”。

### 5.14.2 全局变量表

| 全局变量 | 类型 | 当前作用 |
| -------- | ---- | -------- |
| `Version` | `const string` | 项目版本常量，当前值为 `1.0.0` |
| `Flags` | `*FlagRecord` | 保存启动参数中需要跨模块复用的信息，当前主要是配置文件路径 |
| `Config` | `*conf.Config` | 全局运行时配置，供路由、业务逻辑、中间件读取 |
| `Logger` | `*logrus.Logger` | 全局日志器 |
| `DB` | `*gorm.DB` | MySQL 数据库连接 |
| `Redis` | `*redis.Client` | Redis 客户端 |
| `KafkaMysqlClient` | `*kafka_service.KafkaMysqlClient` | Kafka 同步客户端 |
| `ESClient` | `*elasticsearch.Client` | Elasticsearch 客户端 |
| `ImageCaptchaStore` | `base64Captcha.Store` | 图形验证码内存存储 |

### 5.14.3 初始化顺序与依赖关系

#### 初始化顺序

`global` 中的变量不是懒加载的，而是在 `main.go` 启动时按顺序初始化：

1. `global.Flags`
2. `global.Config`
3. `global.Logger`
4. `global.Redis`
5. `global.KafkaMysqlClient`
6. `global.DB`
7. `global.ESClient`

后续 `flags.Run()`、`core.InitMySQLES()`、`cron_service.Cron()`、`router.Run()` 以及各业务模块，都是直接消费这些全局对象。

#### 设计思路

- 当前项目没有引入依赖注入框架，也没有在每层函数签名中层层透传 `db/logger/config`。
- 选择 `global` 的直接原因，是降低各层接线成本，让 handler、中间件、服务函数都能直接拿到核心依赖。
- 对于这个体量的项目，这种方式开发效率更高，尤其适合：
  - 路由层快速挂业务
  - 中间件复用统一配置
  - 服务层复用数据库 / Redis / ES 连接

### 5.14.4 关键实现思路

#### 1. `global.Config` 是运行时配置镜像

- `core.ReadCfg()` 会把 YAML 文件读入 `global.Config`；
- 业务代码读取站点开关、Gin 模式、JWT 配置、邮件配置时，都是直接读取这份内存对象；
- 当管理员修改站点配置时，代码会先更新 `global.Config`，再通过 `core.SetCfg(global.Config, &global.Flags.File)` 把内容回写到启动时指定的 YAML 文件。

这说明 `global.Config` 当前不是只读快照，而是“内存态 + 文件态”双向同步的运行时配置对象。

#### 2. `global.Flags` 的核心价值是保留配置文件路径

`FlagRecord` 当前看起来只有一个 `File` 字段，似乎很薄，但它并不是无意义占位。

它的真实用途至少有两个：

- 启动时记录本次进程到底是用哪份配置文件启动的；
- 后续站点配置修改时，给 `core.SetCfg()` 提供配置文件落盘路径。

也就是说，`global.Flags` 现在承担的是“配置文件来源追踪”能力，而不是完整保存所有命令行参数。

#### 3. `ImageCaptchaStore` 统一了验证码生成和校验的存储后端

- 图形验证码生成接口会把验证码写入 `global.ImageCaptchaStore`
- `CaptchaMiddleware` 校验时也读取同一个 store

当前默认用的是 `base64Captcha.DefaultMemStore`。这个设计简单直接，不需要单独建表或 Redis，但也意味着验证码状态只保存在当前进程内存里。

#### 4. 全局客户端复用连接池

`DB`、`Redis`、`ESClient` 这类对象都属于重量级连接资源。把它们放在 `global` 中单例复用，避免了每个请求临时创建连接，也让中间件和业务层能共享同一套连接池配置。

### 5.14.5 使用场景

| 使用场景 | 依赖的全局对象 | 说明 |
| -------- | -------------- | ---- |
| 路由启动 | `global.Config` | 读取 Gin 运行模式和监听地址 |
| 日志输出 | `global.Logger` | 控制器、服务、中间件统一输出日志 |
| 数据库读写 | `global.DB` | 所有 GORM 查询、事务、软删 / 级联逻辑都依赖它 |
| 缓存与限流 | `global.Redis` | 点赞计数、验证码、聊天限流等能力依赖 Redis |
| ES 同步 | `global.ESClient` | 搜索索引相关操作复用同一个 ES 客户端 |
| 站点配置回写 | `global.Config + global.Flags` | 更新配置后写回启动时使用的 YAML |
| 图形验证码 | `global.ImageCaptchaStore` | 验证码生成与校验共享内存存储 |

### 5.14.6 当前实现边界

- 当前是典型的包级全局状态方案，依赖访问简单，但耦合也更强，单测需要手动重置全局变量。
- 初始化顺序有隐式约束。比如某些逻辑如果在 `global.Logger` 或 `global.DB` 未完成初始化前就被调用，会直接出错。
- `global.Config` 在运行期间会被站点配置接口直接修改，但当前没有额外的读写锁保护；这在现阶段问题不大，但本质上仍然属于共享可变状态。
- `FlagRecord` 目前只保存 `File`，没有把 `DB`、`ES`、`Type`、`Sub` 等启动参数也一起保留下来，因此它更像“配置文件路径记录器”，而不是完整启动上下文。
- `Version` 已定义在 `global` 中，但当前命令行还没有真正把它输出给用户。



## 文章AI





## 数据统计





## 其他工具模块

### IP 地址归属地解析

- 基于 `ip2region` 离线库实现，数据文件位于 `init/ipbase/ip2region_v4.xdb` 与 `init/ipbase/ip2region_v6.xdb`。
- 项目启动时由 `core.InitIPDB()` 初始化 IPv4/IPv6 搜索器。
- 运行时通过 `core.GetIpAddr(ip)` 解析登录日志、操作日志中的地区信息。

| 能力 | 当前状态 | 说明 |
| ---- | -------- | ---- |
| IPv4 地址解析 | 已实现 | 返回“省份·城市”或“国家·地区” |
| IPv6 地址解析 | 已实现 | 同上 |
| 内网地址识别 | 已实现 | 内网地址直接返回“内网地址” |
| 异常兜底 | 已实现 | 查询失败返回“未知地址” |

### 参数校验与错误翻译

- 请求参数绑定主要通过 `BindJson`、`BindQuery`、`BindUri` 中间件完成。
- 统一错误响应通过 `common/res` 输出。
- `FailWithError` 会尽量把校验类错误翻译成更可读的中文提示。

### JWT 与黑名单能力

- JWT 使用 `HS256` 签名，核心载荷包含 `user_id`、`role`、`username`。
- 接口读取 token 的位置为：
  - Header：`token`
  - Query：`token`
- Redis 黑名单能力已经存在，可支持主动失效、设备级失效等扩展场景。

### 中间件能力

| 中间件 | 当前状态 | 说明 |
| ------ | -------- | ---- |
| `LogMiddleware` | 已实现 | 记录请求体、响应体、响应头并落日志库 |
| `AuthMiddleware` | 已实现 | 校验登录态与 token 黑名单 |
| `AdminMiddleware` | 已实现 | 校验管理员权限 |
| `CaptchaMiddleware` | 已实现 | 校验图片验证码 |
| `EmailVerifyMiddleware` | 已实现 | 校验邮箱验证码 |
| `BindJson/BindQuery/BindUri` | 已实现 | 统一参数绑定，减少控制器样板代码 |

### 设计补充

- 中间件层的设计目标，是把“横切逻辑”从业务 handler 中剥离出去。
- 当前项目已经较明显地按这个方向实现了几类公共能力：
  - 认证鉴权
  - 验证码校验
  - 日志记录
  - 请求参数绑定
- 这样做的直接收益是接口函数更薄，业务逻辑更集中在 `api` 和 `service` 层，后期要补测试或重构时边界更清晰。



## 预留模块

以下主题在文档中保留，但当前仓库没有完整路由与业务实现，不应再按“已支持”描述：

| 模块     | 当前状态 | 说明                                   |
| -------- | -------- | -------------------------------------- |
| 文章 AI  | 预留     | 配置层有 `AI` 配置项，但无实际业务 API |
| 数据统计 | 预留     | 暂无独立统计接口                       |



# 六、项目部署

修改 Docker Compose 和 编写 Dockerfile

## 申请证书

### 安装 CertBot

在 Ubuntu 上通过 Snap 包管理器安装 `Let's Encrypt` 的 `Certbot` 工具（用于免费申请和管理 HTTPS 证书）

```bash
apt update
apt install -y snapd
snap install core
snap refresh core
snap install --classic certbot
ln -sf /snap/bin/certbot /usr/bin/certbot
```

### HTTP-01 方式

#### 简介

Let’s Encrypt 会访问你的网站：

```
http://yourdomain.com/.well-known/acme-challenge/xxx
```

如果返回正确内容 → 证明你控制这个域名

必须开放 **80端口**，在 80 端口放一个文件让 CA 访问，80 端口不能被占用（nginx / go 服务要停）

CDN 场景容易失败

#### 实操

进入服务器后台管理，开放 80 和 443 端口，并新增解析记录。

然后，用 `Certbot` 获取证书

```bash
sudo certbot certonly --standalone \
  -d 你的域名 \
  --agree-tos \
  -m 你的邮箱(用于接收失效信息) \
  --non-interactive

# 获取完后会提示证书位置
# Certificate is saved at:
# /etc/letsencrypt/live/你的域名/fullchain.pem
# Key is saved at:
# /etc/letsencrypt/live/你的域名/privkey.pem
```

### DNS-01

#### 简介

Let’s Encrypt 会检查：

```
_acme-challenge.yourdomain.com 的 TXT 记录
```

如果存在指定值 → 验证通过

#### 实操

用 `Certbot` 获取证书

```bash
sudo certbot certonly \
  -d 你的域名 \
  --manual \
  --preferred-challenges dns \
  --agree-tos \
  -m 你的邮箱
```

然后，弹出提示如下：
```bash
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Please deploy a DNS TXT record under the name:

_acme-challenge.你的域名

with the following value:

要填入的检验码

Before continuing, verify the TXT record has been deployed. Depending on the DNS
provider, this may take some time, from a few seconds to multiple minutes. You can
check if it has finished deploying with aid of online tools, such as the Google
Admin Toolbox: https://toolbox.googleapps.com/apps/dig/#TXT/_acme-challenge.你的域名.
Look for one or more bolded line(s) below the line ';ANSWER'. It should show the
value(s) you've just added.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Press Enter to Continue
```

这时，前往你的域名管理后台，添加一条解析记录：

- 域名：_acme-challenge.你的域名
- 类型：TXT
- 记录值：上面提到的`要填入的检验码`

添加完成后，使用访问工具，看看检验码有没有被检测出来：

```bash
# 浏览器访问：
https://toolbox.googleapps.com/apps/dig/#TXT/_acme-challenge.你的域名
```

多次刷新，直到成功拿到对应的检验码，然后回到申请的地方按下 Enter 继续即可（一般要等个几分钟）

```bash
# 获取完后会提示证书位置
# Certificate is saved at:
# /etc/letsencrypt/live/你的域名/fullchain.pem
# Key is saved at:
# /etc/letsencrypt/live/你的域名/privkey.pem
```



### 拿出证书

复制到部署目录

```bash
cp /etc/letsencrypt/live/你的域名/fullchain.pem \
   /opt/myblogx/myblogx_server/init/deploy/nginx/cert/domain.pem

cp /etc/letsencrypt/live/你的域名/privkey.pem \
   /opt/myblogx/myblogx_server/init/deploy/nginx/cert/domain.key

chmod 600 /opt/myblogx/myblogx_server/init/deploy/nginx/cert/domain.key
chmod 644 /opt/myblogx/myblogx_server/init/deploy/nginx/cert/domain.pem
```

**续期**：Let's Encrypt 证书 90 天一续。

后面可以先手动这样做：

```bash
certbot renew
cp /etc/letsencrypt/live/你的域名/fullchain.pem /opt/myblogx/myblogx_server/deploy/nginx/cert/domain.pem
cp /etc/letsencrypt/live/你的域名/privkey.pem /opt/myblogx/myblogx_server/deploy/nginx/cert/domain.key
cd /opt/myblogx/myblogx_server/deploy && docker compose restart blog_web
```

# 七、非功能性需求（忽略此部分，未实际添加到项目中去，暂不更改，后续视开发进度进行补充）

## 并发安全

**TODO**

tx 数据库事务 / 锁

```go
// 1. 数据库层：给 user 表的 username 字段加唯一索引（关键！）
// 2. 代码层：用事务包裹更新逻辑，捕获唯一键冲突错误
tx := global.DB.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 先校验 Count（前置友好提示）
var nameCount int64
if err = tx.Model(&models.UserModel{}).Where("username = ?", cleanUsername).Count(&nameCount).Error; err != nil {
    tx.Rollback()
    res.FailWithError(err, c)
    return
}
if nameCount > 0 {
    tx.Rollback()
    res.FailWithMsg("用户名已被使用", c)
    return
}

// 更新用户信息
userModel.Username = cleanUsername
userModel.UserConfModel.UpdatedUsernameDate = time.Now()
if err = tx.Save(&userModel).Error; err != nil {
    // 捕获唯一键冲突（应对极端并发）
    if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "唯一索引") {
        tx.Rollback()
        res.FailWithMsg("用户名已被使用", c)
        return
    }
    tx.Rollback()
    res.FailWithError(err, c)
    return
}

tx.Commit()
```



## 分布式系统

**TODO**

Redis、jwt、Mysql主从、Es分片

用Kafka通信，拆分日志系统，定时同步系统，ES桥接系统

### 延迟问题

MySQL 主从同步的核心是主库把操作记录到 binlog，从库再复制这些操作。但数据量一大，从库复制会有延迟 —— 刚写完的数据，立刻读从库可能拿到旧数据或无数据。

1. **简单粗暴：写后即读的情况，直接走主库**：

   如果是「编辑文章后预览」「提交表单后立即查看」这类场景，直接让请求走主库，彻底避开从库延迟：

   - 伪代码：读写分离路由逻辑

     ```go
     func queryArticle(ctx context.Context, articleID int64) (*Article, error) {
         // 关键判断：如果是“写后立即读”的请求（如用户刚提交/更新文章），路由到主库
         if ctx.Value("is_after_write") == true {
             return masterDB.QueryArticle(articleID) // 主库读取
         }
         // 普通读请求，路由到从库
         return slaveDB.QueryArticle(articleID)
     }
     ```

   - GORM 提供单次查询走主库的方法：

     ```go
     // 核心：添加 Clauses(dbresolver.Write) 强制使用主库
     var article model.Article
     err := db.Clauses(dbresolver.Write).Take(&article, "id = ?", articleID).Error
     ```

2. **更优解：缓存兜底（主库压力也能省）**：

   - 写流程：

     ```
     主库写入 → 同步更新缓存（Redis） → 返回成功
     ```

   - 读流程：

     ```
     读请求 → 查缓存 → 缓存命中 → 返回数据（无需访问数据库）
            → 缓存未命中 → 读主库（仅这一次） → 回写缓存 → 返回数据
     ```

3. **低成本：延迟读（让从库 “赶得上”）**：

   - 前端层面：用户提交表单后，显示 “处理中”，延迟 500ms 再发起查询请求（比如用 setTimeout）；
   - 后端层面：对非核心读请求，封装 “延迟读” 函数，最多重试 3 次（每次间隔 200ms），直到从库读到最新数据或超时（超时则读主库）。

4. **进阶玩法：加一层中间件**：

   想自动化管理读写分离，则可加一层中间件：`ProxySQL`。

   `ProxySQL` 由 C++ 编写，在众多支持读写分离的中间件中性能最高，同时可实现故障自动切换等功能，但不支持自动分库分表。

   - 配置示例：article 表的 UPDATE/INSERT 后 5 秒内的 SELECT 请求读主库

     ```sql
     INSERT INTO mysql_query_rules (rule_id, match_pattern, destination_hostgroup, delay_before_route)
     VALUES (100, '^SELECT.*FROM article WHERE id = .*', 1, 5000);
     -- hostgroup 1：主库组，hostgroup 2：从库组
     ```

     实际配置肯定不会是这么简单，不然主库压力太大。可以启用 Lua 脚本路由，能读取 SQL 中的 article_id，再查数据库获取该文章的最后更新时间，动态判断路由目标是主库还是从库。

   - 高可用：搭配 `Keepalive`（局域网）或`云厂商负载均衡`，避免单点故障导致崩溃。

### 缓存问题



### 消息队列

安装 Kafka，进行消息处理

### Mysql

#### 读写分离

1. 基础：gorm 配置
2. 进阶：ProxySQL 自动读写分离 + 更高级的功能
3. 更多功能：见 [分库分表] 部分

#### 分库分表

当数据量 / 用户量涨到一定程度，单库单表扛不住，则需要思考怎么来将数据分开存储，保障搜索性能，避免超过硬盘内存。

##### 阶段 1：不用拆（小体量省心）

用户量≤10 万、数据量≤10G → 单库单表直接用，完全够用。

##### 阶段 2：单库分多表（中等体量）

把一个大表拆成多个结构一样的小表，比如 `user` 表拆成 `user_00`、`user_01`…

- 拆分规则：按 ID 哈希最常用（比如 `user_id % 100 = 0` 存 `user_00`）；
- 效果：1 亿用户拆 100 张表，每张表仅 100 万数据，查询速度直接起飞；
- 核心原则：单表数据控制在 100 万～1000 万行，MySQL 跑得贼快。

##### 阶段 3：分库分表（大规模）

单库性能到瓶颈（比如 QPS 超 1 万、数据超 100G）→ 既拆表，又拆库，分散到多台服务器。

- 中小项目：用 GORM 分表插件（`gorm.io/plugin/sharding`），实现单个Mysql应用内分库分表，不用部署额外服务，代码里配好就行；
- 大型项目：选中间件 ——MyCat2（支持多语言，功能全但性能稍弱）、Vitess（Go 语言开发，性能高，云原生友好）。

### Redis

1

如果要部署多个后端实例，要解耦 Redis 缓存刷库逻辑，单独开个程序处理 Redis 刷库。

### Elasticsearch

1

### 分布式锁

1

## 微服务架构

**TODO**

拆分为各种服务，文件架构更好看

可利用消息队列？

## 防 DDos

- 图形验证码是一个容易被利用的Ddos接口，因为不用验证权限，且生成图片需耗费一定时间。
- Cloudfare 免费 cdn 代理：带免费 ddos 防护，但是国内用户使用可能访问速度较慢。
  - 可加强防护，在 WAF 上加入 JS 质询，判断是否真实浏览器用户（但这样要用户等待一段时间）。
  - 访问速度较慢解决方法：看[视频](https://www.bilibili.com/video/BV1SM4m1176E) 使用 [WeTest.vip](https://www.wetest.vip/) 。
- Ng

## 消息队列设计

### 一、整体架构（推荐的标准形态）

#### 1) 写操作：Command Queue（强烈推荐）

**API 服务**不直接写库，而是：

- 生成一个“写库命令消息”（例如 `CreateOrder` / `UpsertUserHistory`）
- 写入 MQ
- 返回给前端：`202 Accepted + request_id`（或同步返回一个“已受理”的状态）

**DB Worker（消费者）**

- 消费消息
- 用 GORM 执行事务写库
- 记录处理结果（成功/失败、错误原因）
- 需要的话再发一个“完成事件”或更新一个“任务状态表/缓存”

> 这一套的本质是：**把写库从同步 RPC 变成异步消息驱动**。

#### 2) 读操作：通常不走队列（除非你做 CQRS）

读数据库一般还是：

- API -> DB / Cache 直接查
- 或者走 Elasticsearch / Redis 的读模型

如果你“读也要队列化”，那通常意味着 **CQRS + Read Model**：

- 写入走 MQ
- Worker 写库后再发事件
- 另一个 Worker 更新“读库/ES/Redis”供查询

### 二、消息设计（别上来就传 SQL）

消息里不要传 “SQL 字符串”，而是传**业务命令**（Command）：

```
{
  "id": "uuid",
  "type": "UpsertUserArticleHistory",
  "ts": 1700000000,
  "payload": {
    "user_id": 123,
    "article_id": 456,
    "viewed_at": "2026-02-26T18:00:00+08:00"
  }
}
```

#### 必备字段

- `id`：幂等用（极关键）
- `type`：路由到不同 handler
- `payload`：业务参数
- `ts`：审计/排障

### 三、DB Worker 用 GORM 消费并写库（核心落地要点）

#### 1) 幂等（必须做）

消息队列天然会出现：

- **重复投递**（at-least-once）
- **重试**
   所以你的写库必须幂等。

常用两种做法：

**A. 消息去重表（通用，强推荐）**
 建表 `consumed_messages`：

- `message_id` 唯一索引
- `consumed_at`
   Worker 每次处理前先插入（或事务内插入），插不进去说明处理过，直接 ack。

**B. 业务幂等（你之前问的 Upsert 就是）**
 例如“浏览历史”：唯一键 `(user_id, article_id)`，用 `ON CONFLICT / ON DUPLICATE KEY UPDATE` 做 upsert。

两者可以叠加：**消息级幂等 + 业务级幂等**最稳。

#### 2) 事务边界（要么全成，要么全不成）

Worker 收到一条消息：

- 开事务
- 幂等检查（插入 consumed_messages 或业务 upsert）
- 执行业务写操作
- 提交事务
- 提交 offset / ack 消息

这样才能避免“写库成功但 ack 失败导致重复消费”引发脏数据。

#### 3) 重试与 DLQ（死信队列）

- 临时错误（网络抖动、锁冲突）应重试
- 永久错误（参数不合法、外键不存在）应丢到 DLQ，方便排查
