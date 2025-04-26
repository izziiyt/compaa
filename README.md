[![CI][ci-img]][ci]
[![Go Report Card][go-report-img]][go-report]
[![License: MIT][license-img]][license]

# Why compaa (Component Activity Analyzer)?

`compaa` is simple component activity analyzer designed for secure software development.
You can find maintainance activities and EOLs of dependended modules.
It aims supporting your secure software component maintainance.

# Install

go
```bash
go install github.com/izziiyt/compaa@latest
```

mise
```bash
mise use --global go:github.com/izziiyt/compaa@latest
```

# Example
You can find your software depends on inactive OSS. 
(recommended to use your github token when running for sufficient github api rate limit.)

```bash
GITHUB_TOKEN=${YOUR_GITHUB_TOKEN} compaa ./target/path
./path/example0/Dockerfile
./path/example1/subpath/package.json
./path/example2/Dockerfile
├ WARN: docker.io/library/alpine:3.13 last update is'nt recent (2022-11-10)
./path/example2/subpath/Dockerfile
./path/example3/go.mod
├ WARN: go1.18 is EOL
├ WARN: github.com/pkg/errors is archived
├ WARN: github.com/jinzhu/gorm last push is'nt recent (2023-09-11)
```

# Supported File Format

compaa supports the following file formats:
- Dockerfile (Docker)
- Gemfile (Ruby)
- go.mod (Go)
- package.json (Javascript)
- requirements.txt (Python)

# License
This project is licensed under the MIT License, see the LICENSE file for details.

[ci]: https://github.com/izziiyt/compaa/actions/workflows/ci.yaml
[ci-img]: https://github.com/izziiyt/compaa/actions/workflows/ci.yml/badge.svg
[go-report]: https://goreportcard.com/report/github.com/izziiyt/compaa
[go-report-img]: https://goreportcard.com/badge/github.com/izziiyt/compaa
[license]: https://opensource.org/licenses/MIT
[license-img]: https://img.shields.io/badge/License-MIT-yellow.svg
