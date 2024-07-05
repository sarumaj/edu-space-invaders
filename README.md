# space-invaders

this is an example project to showcase the game development using the Web Assembly Framework in Go.
Web Assembly allows us to develop web-frontend applications in a runtime like Go.

[![Gameover](assets/gameover.png)](assets/gameplay.mp4)

## Setup

To setup similar project follow following steps:

1. Create GitHub repository.
2. [Install](https://github.com/git-guides/install-git) git CLI and [authenticate](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github) it.
3. Clone your repository:
   ```
   git clone https://github.com/[username]/[repository name]
   cd [repository name]
   ```
4. Initialize new Go module: `go mod init github.com/[username]/[repository name]`, where `github.com/[username]/[repository name]` will be your module name.
5. Start coding. Additional libraries can ben added using `go get [module name]`. Use `go mod tidy` if necessary.
6. Define unit tests and execute: `go test -v ./...`
7. Generate Assembly files and goodies for the distribution package: `go generate ./...`
8. Execute: `go run [program entrypoint file]`
9. Build: `go build [program entrypoint file]`
10. Utilize version control:
11. Status check: `git status`
12. Pull: `git pull`
13. Stage and commit:
    ```
    git add .
    git commit -m "[your commit message goes here]"
    ```
14. Push: `git push`
15. Advanced usage:
    1. Create a temporary branch: `git checkout -b [branch name]`
    2. Pull, stage, commit
    3. Push: `git push --set-upstream origin [branch name]`
    4. Create pull request and merge it through the web interface ([github.com](github.com))

## Application structure

- [game server main.go](cmd/space-invaders/main.go)
- [module file go.mod](go.mod)
- [source directory](src)
  - [package pkg](src/pkg)
    - [package config](src/pkg/config)
      - [code file const.go](src/pkg/config/const.go)
      - [code file js.go](src/pkg/config/js.go)
      - [code file os.go](src/pkg/config/os.go)
    - [package handler](src/pkg/handler)
      - [code file handler.go](src/pkg/handler/handler.go)
      - [code file handler_js.go](src/pkg/handler/handler_js.go)
      - [code file handler_os.go](src/pkg/handler/handler_os.go)
    - [package objects](src/pkg/objects)
      - [code file bullet.go](src/pkg/objects/bullet.go)
      - [code file bullets.go](src/pkg/objects/bullets.go)
      - [code file enemies.go](src/pkg/objects/enemies.go)
      - [code file enemy.go](src/pkg/objects/enemy.go)
      - [code file enemylevel.go](src/pkg/objects/enemylevel.go)
      - [code file enemytype.go](src/pkg/objects/enemytype.go)
      - [code file position.go](src/pkg/objects/position.go)
      - [code file size.go](src/pkg/objects/size.go)
      - [code file spaceship.go](src/pkg/objects/spaceship.go)
      - [code file spaceshiplevel.go](src/pkg/objects/spaceshiplevel.go)
      - [code file spaceshipstate.go](src/pkg/objects/spaceshipstate.go)
    - [directory static](src/static)
      - [static file favicon.ico](src/static/favicon.ico)
      - [static file index.html](src/static/index.html)
      - [static file style.css](src/static/style.css)
      - [static file wasm.js](src/static/wasm.js)
    - [build script build.sh](src/build.sh)
    - [game entrypoint main.go](src/main.go)

The script [build.sh](src/build.sh) is meant to compile the web assembly package (main.wasm) and create a distribution package [dist](dist).
The [game server](cmd/space-invaders/main.go) serves the files from the distribution package using the web assembly. The files can be served in any other runtime than Go.
Some code components are meant to be compiled only for the JS WASM architecture (e.g. [js.go](src/pkg/config/js.go) and [handler_js.go](src/pkg/handler/handler_js.go)).
To be able to compile the code for other targets and to run tests against it, some mock-ups haven been defined (e.g. [os.go](src/pkg/config/os.go) and [handler_os.go](src/pkg/handler/handler_os.go)). The heart of the web application is the JavaScript script building the bridge between the WASM package: [wasm.js](src/static/wasm.js) and our static web page: [index.html](src/static/index.html).

## Furter reading

- [WebAssembly](https://go.dev/wiki/WebAssembly)
