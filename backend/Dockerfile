# ==========================================
# 第一阶段：构建 VSentry 后端主程序
# ==========================================
FROM golang:1.25-alpine AS builder

WORKDIR /build

# ✅ 关键修复 1：安装所有 CGO 依赖 (不仅是 sqlite，还需要 libpcap)
RUN apk update && apk add --no-cache gcc musl-dev sqlite-dev libpcap-dev

# 1. 缓存依赖，加速后续构建
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 2. 拷贝后端所有代码并编译单体二进制
COPY backend/ ./backend/
WORKDIR /build/backend
RUN go mod tidy
# 开启 CGO for SQLite & PCAP, Trimpath 并精简符号表
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags "-s -w" -o vsentry .


# ==========================================
# 第二阶段：构建最终运行时环境
# ==========================================
FROM golang:1.25-alpine

WORKDIR /app

# ✅ 关键修复 2：运行时同样需要这些编译工具和运行时库
# tzdata   - 时区
# sqlite   - SQLite 运行时支持
# libpcap  - VSentry 主程序调用 PCAP 时需要的动态链接库 .so
# libpcap-dev, gcc, musl-dev - 因为你的应用要在运行时动态 build agent，如果 agent 也需要 cgo，这些必须带着
RUN apk update && apk add --no-cache \
    tzdata \
    sqlite \
    libpcap \
    libpcap-dev \
    gcc \
    musl-dev

# 1. 拷贝编译好的后端主程序
COPY --from=builder /build/backend/vsentry ./vsentry

# 2. 拷贝动态编译 Agent 所需的"源码 SDK"
COPY backend/go.mod backend/go.sum ./
COPY backend/pkg/ ./pkg/
COPY backend/cmd/ ./cmd/

# 3. 预下载 Agent 的编译依赖 (现在这里不会报错了，因为有 gcc 和 pcap 头文件)
RUN go mod download

# 4. 拷贝静态配置
COPY backend/config/ ./config/

EXPOSE 8088

# 挂载数据卷以持久化 GOCACHE
ENV GOCACHE=/app/data/.gocache

CMD ["./vsentry"]