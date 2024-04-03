[![CI][ci-img]][ci]
[![Go Report Card][go-report-img]][go-report]
[![License: MIT][license-img]][license]

# compaa

compaa is component activity analyzer designed for secure software development. You can find maintainance activities and EOLs of dependended OSS.
It aims supporting your secure software component maintainance.

# Install

```shell
go intsall github.com/izziiyt/compaa@v0.2.3
```

# Sample
You can find your sofware depends on inactive OSS. (recommended to set your github token when running for sufficient github api rate limit.)
```shell
compaa -t ${YOUR_GITHUB_TOKEN} ./path
./path/sample0/Dockerfile
./path/sample1/subpath/package.json
./path/sample2/Dockerfile
├ WARN: docker.io/library/alpine:3.13 last update isnt recent (2022-11-10 20:55:35.397295 +0000 UTC)
./path/sample2/subpath/Dockerfile
./path/sample3/go.mod
├ WARN: go1.18 is EOL
├ WARN: github.com/pkg/errors is archived
├ WARN: github.com/jinzhu/gorm last push isnt recent (2023-09-11 08:16:54 +0000 UTC)
```

# License
This project is licensed under the MIT License, see the LICENSE file for details.

# Supports

- Dockerfile
- Gemfile
- go.mod
- package.json
- requirements.txt

# Note

It may occur breaking changes in minor update.

[ci]: https://github.com/izziiyt/compaa/actions/workflows/ci.yaml
[ci-img]: https://github.com/izziiyt/compaa/actions/workflows/ci.yml/badge.svg
[go-report]: https://goreportcard.com/report/github.com/izziiyt/compaa
[go-report-img]: https://goreportcard.com/badge/github.com/izziiyt/compaa
[license]: https://opensource.org/licenses/MIT
[license-img]: https://img.shields.io/badge/License-MIT-yellow.svg
