# Make rules
## Overall
Use ENV to control all Makefile rules. Implement all Makefile rules in shell scripts.

NOTE: All make rules are not dependent on the other rules. 

-   If you want to make container, you must make build firstly.
-   If you want to make push, you must make container firstly.

A general use case is `make build container push`.

## Prerequisite
- Place your project in `$GOPATH`
- Place `make-rules` under `hack` dir in your project

## Install 

As git submodule (recommended)

```
cd your/project
mkdir -p hack
cd hack && git submodule add https://github.com/zoumo/make-rules
cd make-rules && sh install.sh
```

## Dir Convention

The build system follows the dir conventions:

```
├── bin
│   ├── darwin_amd64
│   └── linux_amd64
├── build
│   ├── server
│   └── worker
├── cmd
│   ├── server
│   └── worker
├── hack
│   └── make-rules
│       ├── lib
│       └── entrypoint
├── pkg
│   └── version
└── test
```
- `bin` contains build outputs, should be ignored in `.gitignore`.
    - `${GOOS}_${GOARCH}` contains platform-based build outputs.
- `build` contains several directories must holding Dockerfile.
- `cmd` contains main packages, each subdirecoty of cmd is a main package.
- `hack` contains scripts used to manage this repository.
    - `make-rules`
        - `lib` contains scripts library.
        - `make-rules` contains scripts which implement Makefile rules.
- `pkg/version` contains version code to represent project's version. Version will be filled in during the compilation phase.

## Help Message

All Makefile rules should print usage messages when `${HELP} == y`

```bash
HELP=y make all
```

## Color Log

| Name      | Usage                                  | Type   | Default |
| --------- | -------------------------------------- | ------ | ------- |
| COLOR_LOG | If set true, the log will be colorized | string | true    |

```bash
COLOR_LOG=true make all
```

## Version

The new version system based on git work tree status.

| Name               | Usage                                    | Type   | Default |
| ------------------ | ---------------------------------------- | ------ | ------- |
| PRJ_GIT_COMMIT     | The git commit id corresponding to this source code | string |         |
| PRJ_GIT_TREE_STATE | "clean" indicates no changes since the git commit id<br />"dirty" indicates source code changes after the git commit id<br />"archive" indicates the tree was produced by 'git archive' | string |         |
| PRJ_GIT_VERSION    | "vX.Y" used to indicate the last release version. Comes from `git describe` , based on `git tag`. | string |         |
| PRJ_GIT_REMOTE     | The git remote origin url.               | string |         |
| VERSION            | alias for PRJ_GIT_VERSION                | string |         |

### User defined version

```bash
VERSION=v0.0.4 make all
```

>   *If your git repo contains changes not staged, a `-dirty` suffix will always be appended to final version.*

### Usage

#### Common

1.  commit all staging changes, make the git tree `clean`.
2.  use `git tag` add a new semantic version tag, e.g `v0.0.3`
3.  the git version and docker tag will be `v0.0.3`

#### User defined version

1.  VERSION=v0.0.4 make all
2.  if you are working in a `dirty` git tree, the version will be `v0.0.4-dirty`
3.  if you are working in a `clea` git tree, the version will be `v0.0.4`

#### Dirty

1.  working in a `dirty` git tree, the latest tag is `v0.0.2`
2.  no new commits
3.  make build
4.  the git version and docker tag will be `v0.0.2-dirty`

#### Commits Tracking

1.  working in a `clean` git tree, the latest tag is `v0.0.2`
2.  commit 1 new change
3.  make build
4.  the git version will be `v0.0.2-1+1b4531e1acf800`, the docker tag will be `v0.0.2-1-1b4531e1acf800`
5.  that means since the latest tag, there is **one** new commit, and the latest commit id is `1b4531e1acf800`

## Go

Persistent Options:

| Name               | Usage                                    | Type   | Default                  |
| ------------------ | ---------------------------------------- | ------ | ------------------------ |
| LOCAL_BUILD        | If set true, project will be built on local machine, otherwise, built in docker. | string | false                    |
| GO_ONBUILD_IMAGE   | Porject will be built in the image when `${LOCAL_BUILD} != true` | string | golang:1.9.2-alpine3.6   |
| GO_BUILD_PLATFORMS | The project will be built for these platforms. | array  | linux/amd64 darwin/amd64 |
| GOFLAGS            | Extra flags to pass to 'go' when building. | string |                          |
| GOLDFLAGS          | Extra linking flags passed to 'go' when building. | string |                          |
| GOGCFLAGS          | Additional go compile flags passed to 'go' when building. | string |                          |

### Build
Supported rules:
- make all
- make $()
- make build
- make build-local
- make build-in-container

Other Options:

| Name                | Usage                                    | Type   | Default      |
| :------------------ | ---------------------------------------- | ------ | ------------ |
| GO_BUILD_TARGETS    | All pre-defined directory names of targets for go build. e.g `cmd/server` | array  | user defined |
| GO_STATIC_LIBRARIES | Determine which go build targets should use `CGO_ENABLED=0`. | array  | user defined |
| WHAT                | Directory names to build.  If any of these directories has a main package, the build will produce executable files under bin/. e.g `cmd/server` | string |              |

Usage:
```bash
# make all targets for all platforms
make all
make build

# make all targets for linux/amd64 
make all GO_BUILD_PLATFORMS=linux/amd64

# make cmd/server for linux/amd64
make all WHAT=cmd/server GO_BUILD_PLATFORMS=linux/amd64
make cmd/server GO_BUILD_PLATFORMS=linux/amd64

# make cmd/server for linux/amd64 in docker container
make build-in-container WHAT=cmd/server GO_BUILD_PLATFORMS=linux/amd64
LOCAL_BUILD=false make all WHAT=cmd/server GO_BUILD_PLATFORMS=linux/amd64
LOCAL_BUILD=false make cmd/server GO_BUILD_PLATFORMS=linux/amd64

# make cmd/server for linux/amd64 on local
make build-local WHAT=cmd/server GO_BUILD_PLATFORMS=linux/amd64
LOCAL_BUILD=true make build WHAT=cmd/server GO_BUILD_PLATFORMS=linux/amd64
LOCAL_BUILD=true make cmd/server GO_BUILD_PLATFORMS=linux/amd64
```

### Test

#### Unit tetst

Supported rules:
- make unittest

Other options

| Name               | Usage                                    | Type  | Default |
| ------------------ | ---------------------------------------- | ----- | ------- |
| GO_TEST_EXCEPTIONS | Go test will ignore the pkg under exceptions dirs.<br />` vendor test tests scripts hack` are always be skipped. | array |         |

Usage:

```bash
make unittest
```

#### E2E test

TODO

## Container

Supported rules:
- make container
- make push

Options:

| Name                 | Usage                                    | Type   | Default |
| -------------------- | ---------------------------------------- | ------ | ------- |
| DOCKER_BUILD_TARGETS | All pre-defined directory names of targets for docker build. e.g `build/server` | array  |         |
| DOCKER_REGISTRIES    | Docker registries to push                | array  |         |
| DOCKER_IMAGE_PREFIX  | Docker image prefix.                     | string |         |
| DOCKER_IMAGE_SUFFIX  | Docker image suffix.                     | string |         |
| DOCKER_FORCE_PUSH    | Force pushing to override images in remote registries | string | true    |
| WHAT                 | Directories containing Dockerfile        | string |         |

Usage: 

```bash
# build all docker images
make container

# push all docker images
make push

# build and push build/server docker iamge
make container push WHAT=build/server
```

