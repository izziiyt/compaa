# Why compaa

compaa is the component activity analysis tool for software security.
It aims supporting your software component analysis.
Some tools supports auto update functions, but you may find it sometimes wont' make it because some library is archived, inactive maintained etc.

# Install

```shell
go intsall github.com/izziiyt/compaa@v0.2.1
```

# Sample
You should set your github token for sufficient github api rate limit.
```shell
compaa -t ${YOUR_GITHUB_TOKEN} ./path
./path/sample0/Dockerfile
./path/sample1/subpath/package.json
./path/sample2/Dockerfile
├ WARN: docker.io/library/alpine:3.13 last update isn't recent (2022-11-10 20:55:35.397295 +0000 UTC)
./path/sample2/subpath/Dockerfile
./path/sample3/go.mod
├ WARN: go1.18 is EOL
├ WARN: github.com/pkg/errors is archived
├ WARN: github.com/jinzhu/gorm last push isn't recent (2023-09-11 08:16:54 +0000 UTC)
```

# License
This project is licensed under the MIT License, see the LICENSE file for details.

# Supports

- go.mod
- package.json
- Dockerfile
- requirements.txt
- Gemfile

# Note

It may occur breaking changes in minor update.
