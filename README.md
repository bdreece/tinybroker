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

tinybroker is a message broker, which implements the pub/sub model, written in Go. The message specification for tinybroker is written using Google Protobufs, so as to port easily to many languages.

---

## Getting Started

### Downloading and Installing

Downloading tinybroker is as simple as:

```console
$ go install github.com/bdreece/tinybroker@latest
```

If you would like to mess around with the library sources as well:

```console
$ go get github.com/bdreece/tinybroker/tb
```

### Running

Once you've installed tinybroker, the executable should be in your `$GOPATH/bin` directory. This can be executed as `tinybroker`, assuming you've configured go correctly.

---

## Usage

### API Reference

The API reference is currently under development, but I plan to use some amalgam of Godoc and Doxygen to generate from source code.

### tinybrokerd

The `tinybroker` CLI is still under development. Check back here for updated regarding its usage

---

## Future Plans

In the future, I may make the Broker library API a bit more configurable (i.e. byte pre- and post-processiong, etc.), and I plan to flesh out the broker daemon CLI in the coming weeks.
