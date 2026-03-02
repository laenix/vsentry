# ==========================================
# 第一阶段：构建 VSentry 后端主程序
# ==========================================
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 1. 缓存依赖，加速后续构建
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 2. 拷贝后端所有代码并编译单体二进制
COPY backend/ ./backend/
WORKDIR /build/backend
# 开启 Trimpath 并精简符号表
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o vsentry .

# ==========================================
# 第二阶段：构建最终运行时环境 (包含 Go 工具链)
# ==========================================
FROM golang:1.23-alpine

WORKDIR /app

# 安装时区数据，保证日志时间戳准确
RUN apk add --no-cache tzdata

# 1. 拷贝编译好的后端主程序
COPY --from=builder /build/backend/vsentry ./vsentry

# 2. 【核心变更】拷贝动态编译 Agent 所需的“源码 SDK”
# 因为 Agent 导入了 github.com/laenix/vsentry/pkg/ocsf，
# 所以必须保留 go.mod 和 pkg 目录，否则容器内 go build 会报找不到模块。
COPY backend/go.mod backend/go.sum ./
COPY backend/pkg/ ./pkg/
COPY backend/cmd/ ./cmd/

# 3. 预下载 Agent 的编译依赖到容器本地缓存，让前端点击 Build 时瞬间完成
RUN go mod download

# 4. 拷贝静态配置 (保持不变)
COPY backend/config/ ./config/

EXPOSE 8088

# 挂载数据卷以持久化 GOCACHE
ENV GOCACHE=/app/data/.gocache

CMD ["./vsentry"]