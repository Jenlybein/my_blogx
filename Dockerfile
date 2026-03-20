# 阶段 1：基于官方的golang:alpine镜像创建构建环境（alpine是轻量级Linux发行版）
FROM golang:alpine AS builder

# 禁用CGO（C语言交互），让编译出的Go程序是纯静态可执行文件，能在极简的alpine中运行
ENV CGO_ENABLED 0

# 配置国内镜像，加速依赖下载（direct表示找不到的依赖直接从官方拉）
#ENV GOPROXY direct
ENV GOPROXY https://goproxy.cn,direct

# 修改alpine的软件源为阿里云镜像，加速后续apk命令的包下载
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置工作目录为/build（后续命令都在这个目录执行）'
WORKDIR /build

# 将当前宿主机目录下的所有文件复制到容器的/build目录（即Go项目代码）
ADD . .

# 编译Go代码，生成名为main的可执行文件（输出到/build/main）
RUN go build -o main

# 阶段 1：基于极简的alpine镜像创建运行环境，抛弃构建阶段的Go编译环境，减小镜像体积
FROM alpine

# 设置运行时的工作目录为/app
WORKDIR /app

# 从第一阶段（builder）的/build目录复制编译好的main可执行文件到当前镜像的/app目录
COPY --from=builder /build/main /app

# 安装时区数据（tzdata），解决Go程序运行时时区不正确的问题（比如默认UTC，装了才能用CST/Asia/Shanghai）
RUN apk add tzdata

# 容器启动时执行的命令：运行/app/main可执行文件
CMD ["./main"]