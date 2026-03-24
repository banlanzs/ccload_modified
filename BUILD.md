# 构建指南

## 推荐构建流程（包含前端回归测试）

为防止 UI 功能回退，建议在构建前先运行前端回归测试：

```bash
cd ccLoad

# 1. 运行前端回归测试
node --test web/assets/js/channels-render.test.js

# 2. 测试通过后再构建
go build -tags go_json -o ccload_modified.exe .
```

或使用链式命令（测试失败自动中止构建）：

```bash
node --test web/assets/js/channels-render.test.js && go build -tags go_json -o ccload_modified.exe .
```

## 快速构建（跳过前端测试）

```bash
cd ccLoad
go build -tags go_json -o ccload_modified.exe .
```

> ⚠️ 注意：跳过前端测试可能导致 UI 功能缺失未被发现。

## 带版本信息构建

```bash
cd ccLoad
go build -tags go_json -ldflags "\
  -X ccLoad+ccr/internal/version.Version=$(git describe --tags --always) \
  -X ccLoad+ccr/internal/version.Commit=$(git rev-parse --short HEAD) \
  -X 'ccLoad+ccr/internal/version.BuildTime=$(date '+%Y-%m-%d %H:%M:%S %z')' \
  -X ccLoad+ccr/internal/version.BuiltBy=$(whoami)" -o ccload_modified.exe .
```

## 运行测试

### 前端回归测试

```bash
cd ccLoad
node --test web/assets/js/channels-render.test.js
```

### 后端测试

```bash
cd ccLoad
go test -tags go_json ./internal/... -v
```

## 运行程序

```bash
./ccload_modified.exe
```

环境变量配置在 `.env` 文件中。

## 注意事项

- **必须**使用 `-tags go_json` 标签
- Windows 下可执行文件扩展名为 `.exe`
- 首次运行会自动创建数据库和配置文件
- 建议在构建前运行前端回归测试，防止 UI 功能丢失
- 前端测试依赖 Node.js（需本机可用 `node` 命令）
