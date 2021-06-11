# KitA Log

This is the logging module of the Go-KitA kit.
It provides a simple logging interface which is easy to use and implement.
It also provides a usable implementation based on the `log` package of the Go
SDK.

The following logging frameworks are planned to be adapted:
- [ ] Uber [Zap](https://github.com/uber-go/zap) -
  adapt with a module named zap-log
- [ ] [logrus](https://github.com/sirupsen/logrus) -
  adapt with a module named logrus-log

## Key Feature
- [x] Print message via the Print-like function family: `Print`,`Printf`,`Println`
- [x] Support key/value pairs metadata.
- [x] Support named loggers.
- [x] Support level logging.
- [x] Support dynamic runtime logging level control by logger names.
- [X] Support `context.Context` and extracting value from it.

[comment]: <> (## Architecture)

[comment]: <> (## Usage)

<!-- 描述如何使用该项目 -->

## Authors
- dowenliu-xyz <hawkdowen@hotmail.com>

## License
KitA Log is licensed under the MIT.
See [LICENSE](LICENSE) for the full license text.