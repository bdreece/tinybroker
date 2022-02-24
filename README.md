# tinybroker

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/bdreece/tinybroker/Go)
![Lines of code](https://img.shields.io/tokei/lines/github/bdreece/tinybroker)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/bdreece/tinybroker)

A simple message broker, written in Go

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
  - [Downloading and Installing](#downloading-and-installing)
  - [Running](#running)
- [Usage](#usage)
  - [API Reference](#api-reference)
  - [tinybrokerd](#tinybrokerd)
- [Future Plans](#future-plans)

---

## Overview

tinybroker is a message broker, which implements the pub/sub model, written in Go. Clients can interact with the broker's REST API using standard CRUD conventions on the "/tb/{topic}" endpoints. Messages published to the broker are stored in memory using asynchronous channels.

---

## Getting Started

### Downloading and Installing

Downloading tinybroker is as simple as:

```console
$ go install github.com/bdreece/tinybroker@latest
```

### Running

Once you've installed tinybroker, the executable should be in your `$GOPATH/bin` directory. This can be executed as `tinybroker`, assuming you've configured go correctly.

---

## Usage

```
Usage of ./tinybroker:
  -a string
        Listening address and port (default "127.0.0.1:8080")
  -c int
        Topic queue capacity (default 32)
  -v    Enable verbose output
```

---

## Future Plans

In the future, I may make the broker library API a bit more configurable (i.e. byte pre and post-processiong, etc.).
