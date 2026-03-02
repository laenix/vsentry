# ==========================================
# 第一阶段：构建 VSentry 后端主程序
# ==========================================
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 安装 CGO 依赖 (SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 1. 缓存依赖，加速后续构建
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 2. 拷贝后端所有代码并编译单体二进制
COPY backend/ ./backend/
WORKDIR /build/backend
# 开启 CGO for SQLite, Trimpath 并精简符号表
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags "-s -w" -o vsentry .

# ==========================================
# 第二阶段：构建最终运行时环境
# ==========================================
FROM golang:1.23-alpine

WORKDIR /app

# 安装运行时依赖：时区数据 + SQLite
RUN apk add --no-cache tzdata sqlite

# 1. 拷贝编译好的后端主程序
COPY --from=builder /build/backend/vsentry ./vsentry

# 2. 拷贝动态编译 Agent 所需的"源码 SDK"
COPY backend/go.mod backend/go.sum ./
COPY backend/pkg/ ./pkg/
COPY backend/cmd/ ./cmd/

# 3. 预下载 Agent 的编译依赖
RUN go mod download

# 4. 拷贝静态配置
COPY backend/config/ ./config/

EXPOSE 8088

# 挂载数据卷以持久化 GOCACHE
ENV GOCACHE=/app/data/.gocache

CMD ["./vsentry"]