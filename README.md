[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Log](https://github.com/go-kita/log/actions/workflows/log.ci.yaml/badge.svg)](https://github.com/go-kita/log/actions/workflows/log.ci.yaml)
[![GoDoc](https://pkg.go.dev/badge/github.com/go-kita/log/v3)](https://pkg.go.dev/github.com/go-kita/log/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-kita/log)](https://goreportcard.com/report/github.com/go-kita/log)

Translations: [English](README.md) | [简体中文](README.zh_CN.md)

# Log

[comment]: <> (This is the logging module of the Go-KitA framework.)

[comment]: <> (> Project Go-KitA is inspired by the project [Kratos]&#40;https://github.com/go-kratos/kratos&#41;)

[comment]: <> (> and has a lot of reference to its implementation.)

This module provides a simple logging interface which is easy to use and implement.
It also provides a usable implementation based on the `log` package of the Go
SDK.

The following logging frameworks are adapted:
- [x] Uber [Zap](https://github.com/uber-go/zap) - adapt via [zap-log](https://github.com/go-kita/zap-log) .
- [x] [logrus](https://github.com/sirupsen/logrus) - adapt via [logrus-log](https://github.com/go-kita/logrus-log) .

## Features
- [x] Print message via the Print-like function family: `Print`,`Printf`,`Println`
- [x] Support key/value pairs metadata.
- [x] Support named loggers.
- [x] Support level logging.
- [x] Support dynamic runtime logging level control by logger names.
- [X] Support `context.Context` and extracting value from it.

[comment]: <> (## Usage)

<!-- 描述如何使用该项目 -->

## Authors
- dowenliu-xyz <hawkdowen@hotmail.com>

## License
Log is licensed under the MIT.
See [LICENSE](LICENSE) for the full license text.