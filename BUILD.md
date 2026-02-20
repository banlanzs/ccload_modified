# 构建指南

## 快速构建

```bash
cd ccLoad
go build -tags go_json -o ccload+ccr.exe .
```

## 带版本信息构建

```bash
cd ccLoad
go build -tags go_json -ldflags "\
  -X ccLoad+ccr/internal/version.Version=$(git describe --tags --always) \
  -X ccLoad+ccr/internal/version.Commit=$(git rev-parse --short HEAD) \
  -X 'ccLoad+ccr/internal/version.BuildTime=$(date '+%Y-%m-%d %H:%M:%S %z')' \
  -X ccLoad+ccr/internal/version.BuiltBy=$(whoami)" -o ccload+ccr.exe .
```

## 运行测试

```bash
cd ccLoad
go test -tags go_json ./internal/... -v
```

## 运行程序

```bash
./ccload+ccr.exe
```

环境变量配置在 `.env` 文件中。

## 注意事项

- **必须**使用 `-tags go_json` 标签
- Windows 下可执行文件扩展名为 `.exe`
- 首次运行会自动创建数据库和配置文件
