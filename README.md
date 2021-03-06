# mstreamb0t

A simple bot that notifies when a specific manga is released on MangaStream.

 * [Installation](README.md#installation)
      * [Binaries](README.md#binaries)
      * [Via Go](README.md#via-go)
      * [Running with Docker](README.md#running-with-docker)
 * [Usage](README.md#usage)

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/isaacasensio/mstreamb0t/releases).

#### Via Go

```console
$ go get github.com/isaacasensio/mstreamb0t
```

#### Running with Docker

```console
docker run --restart=unless-stopped -d \
    --name mstreamb0t \
    -e "PUSHBULLET_TOKEN=o.myowbAB6HinGRxVDNyHbBXs98rwqfzrcc2v" \
    isaacasensio/mstreamb0t:0.0.1 \
    --interval 3h \
    --manga-names="Hajime,Dragon Ball"
```

## Usage

```console
$ mstreamb0t -h
mstreamb0t -  A simple bot that notifies when a specific manga is released on MangaStream.

Usage: mstreamb0t <command>

Flags:

  --interval         update interval (ex. 10s, 1m, 3h) (default: 1m0s)
  --once             run once and exit, do not run as a daemon (default: false)
  --manga-names      manga names separated by commas (default: <none>)

Commands:

  version  Show the version information.
```
