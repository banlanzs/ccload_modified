# 开发测试指南

## 快速启动

```bash
# 进入项目目录
cd D:\Documents\ccload+ccr-win64\ccLoad

# 启动开发服务器（版本号显示为 dev）
go run -tags go_json .
```

## 环境配置

### 方式1：使用 .env 文件

在项目根目录创建 `.env` 文件：

```bash
# 最小配置（必需）
CCLOAD_PASS=admin123

# 可选配置
PORT=8080
GIN_MODE=debug
SQLITE_PATH=./data/ccload.db
```

启动命令：
```bash
go run -tags go_json .
```

### 方式2：命令行环境变量

```bash
# 直接启动
CCLOAD_PASS=admin123 GIN_MODE=debug go run -tags go_json .

# 指定端口
CCLOAD_PASS=admin123 PORT=8081 go run -tags go_json .

# 指定测试数据库
CCLOAD_PASS=admin123 SQLITE_PATH=./data/test-dev.db go run -tags go_json .
```

## 访问管理界面

服务启动后，访问：

- **管理后台**：http://localhost:8080/web/channels.html
- **登录密码**：admin123（或你设置的密码）

## 测试功能

### 测试 API Key 标签功能

1. 点击 **"+ 添加渠道"** 按钮
2. 在 API Key 表格中，会看到新增的 **"标签"** 列
3. 每个 Key 可以添加标签（如"公开"、"个人"等备注）
4. 保存后，标签会存储到数据库

### 测试数据

可以使用已有的 SQLite 数据库进行测试：
```bash
# 已有数据的情况
SQLITE_PATH=./data/ccload.db go run -tags go_json .
```

## 构建生产版本

```bash
# 编译为可执行文件
go build -tags go_json -ldflags "
  -X ccLoad+ccr/internal/version.Version=$(git describe --tags --always)
  -X ccLoad+ccr/internal/version.Commit=$(git rev-parse --short HEAD)
  -X 'ccLoad+ccr/internal/version.BuildTime=$(date \"+%Y-%m-%d %H:%M:%S %z\")'
  -X ccLoad+ccr/internal/version.BuiltBy=$(whoami)
" -o ccload+ccr .
```

## 运行测试

```bash
# 所有测试
go test -tags go_json ./internal/... -v

# 带竞态检测
go test -tags go_json -race ./internal/...
```

## 常用命令

| 命令 | 说明 |
|------|------|
| `go run -tags go_json .` | 开发运行 |
| `go build -tags go_json -o ccload+ccr .` | 编译 |
| `go test -tags go_json ./internal/...` | 运行测试 |
| `golangci-lint run ./...` | 代码检查 |

## 端口说明

- **8080**：默认 HTTP 服务端口
- 可以通过 `PORT` 环境变量修改

## 常见问题

### Q: 端口被占用
```bash
# 查看占用进程
netstat -ano | findstr :8080

# 使用其他端口
PORT=8081 go run -tags go_json .
```

### Q: 数据库损坏
```bash
# 删除数据库重新初始化
rm -f ./data/ccload.db
go run -tags go_json .
```

### Q: 编译报错
```bash
# 清理缓存后重试
go clean -cache
go build -tags go_json -o ccload+ccr .
```
