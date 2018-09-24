# goqldom

GraphQL based HTTP service for DOM selections.

## Requirements (once to build)

install [Yarn](https://yarnpkg.com) and [Parcel](https://parceljs.org)

```shell
# yarn global add parcel-bundler
```

install frontend dependencies

```shell
$ yarn install
```

install [Govendor](https://github.com/kardianos/govendor) and [GoReleaser](https://goreleaser.com)

```shell
$ go get -u github.com/kardianos/govendor
$ go get -u github.com/goreleaser/goreleaser
```

synchronize service dependencies

```shell
$ govendor sync
```

## build frontend assets (for every frontend change)

run Parcel to build the static assets

```shell
$ parcel build -d assets index.html
```

build the virtual file system `assets_vfsdata.go` of the static assets

```shell
$ go run service/vfs/generate.go
```

## Run the service

```shell
$ go run service/main.go
```

## Test the release build

```shell
$ ~/go/bin/goreleaser --skip-publish --snapshot --rm-dist
```
