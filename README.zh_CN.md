[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Log](https://github.com/go-kita/log/actions/workflows/log.ci.yaml/badge.svg)](https://github.com/go-kita/log/actions/workflows/log.ci.yaml)
[![GoDoc](https://pkg.go.dev/badge/github.com/go-kita/log/v3)](https://pkg.go.dev/github.com/go-kita/log/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-kita/log)](https://goreportcard.com/report/github.com/go-kita/log)

Translations: [English](README.md) | [简体中文](README.zh_CN.md)

# Log

[comment]: <> (这个 Go-KitA 框架的日志模块。)

[comment]: <> (> Go-KitA 项目受 [Kratos]&#40;https://github.com/go-kratos/kratos&#41; 项目启发，并大量参考了其实现细节。)

本模块提供一套简单易用且易实现的日志接口。同时，也基于Go SDK的 `log` 包提供了一套可用的接口适配实现。

已实现以下日志框架的适配：
- [x] Uber [Zap](https://github.com/uber-go/zap) - 通过 [zap-log](https://github.com/go-kita/zap-log) 适配。
- [x] [logrus](https://github.com/sirupsen/logrus) - 通过 [logrus-log](https://github.com/go-kita/logrus-log) 适配。

## Features

- [x] 使用类 Print 方法进行日志消息体输出： `Print`,`Printf`,`Println`
- [x] 支持输出键值对字段日志元数据。
- [x] 支持命名 logger 。
- [x] 支持日志按级别输出。
- [x] 支持日志级别运行时调整，且可按日志名对各logger分别控制。
- [X] 支持接收 `context.Context` 并从中读取值用于输出。

[comment]: <> (## Usage)

<!-- 描述如何使用该项目 -->

## Contributing

> 欢迎完善英文文档。

> 欢迎完善示例 Examples。

> 欢迎完善 Get Started & Guide & Tutorial。

### Commit messages

我们使用[约定式提交规范](https://www.conventionalcommits.org/zh-hans/v1.0.0/)，提交PR请遵守该规范。

### Documention

请在 Feature 类 PR 中包含必要的文档，bugfix 类 PR 可以忽略文档，但如果 bugfix 导致 BREAKING CHANGE，请明确说明并做必要的文档修改。

代码变更（新增/修改 interface/strut/function/method）中应完善 Go Doc。

### Tests

请在 PR 前完成自测。没有完善测试覆盖的变更都应该被质疑，不论变更内容是新 Feature 还是 bugfix。

## Authors

- dowenliu-xyz <hawkdowen@hotmail.com>

## License

Log is licensed under the MIT. See [LICENSE](LICENSE) for the full license text.