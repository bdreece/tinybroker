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
- [Future Plans](#future-plans)

---

## Overview

tinybroker is a message broker, which implements the pub/sub model, written in Go. Clients can interact with the broker's REST API using standard CRUD conventions on the "/{topic}" endpoints. Authentication is performed via JSON web tokens. Messages published to the broker are stored in memory using asynchronous channels.

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

The command-line usage of tinybroker is as follows:

```
Usage of ./tinybroker:
  -a string
        Listening address and port (default "127.0.0.1:8080")
  -c int
        Topic queue capacity (default 32)
  -v    Enable verbose output
```

Additional parameters (i.e. username/password, JWT HMAC secret) may be passed in as environment variables named `TB_USER`, `TB_PASS`, and `TB_SECRET`, respectively.

### Client-Side

tinybroker exposes its API over HTTP, utilizing standard CRUD conventions for resource access. For a given endpoint, say `/fruits`, a client may manipulate the 'fruits' topic via the following:

- Create topic 'fruits': `POST /fruits`
- Read from topic 'fruits': `GET /fruits`
- Update topic 'fruits': `PUT /fruits`
- Delete topic 'fruits': `DELETE /fruits`

Furthermore, data may be passed along to the broker using the multipart form content type under the key: `TB_DATA`

In order to help illustrate proper broker requests, I've added the following valid `curl` commands for a local tinybroker instance (given the `TB_USER` and `TB_PASS` environment variables have been set to 'user' and 'pass', respectively):

- Request:  `curl -F "TB_USER=user" -F "TB_PASS=pass" localhost:8080/login`
- Response: `<YOUR_JWT_HERE>`
- Request:  `curl --oauth2-bearer "<YOUR_JWT_HERE>" -F "TB_DATA=apple" localhost:8080/fruits`
- Request:  `curl --oauth2-bearer "<YOUR_JWT_HERE>" -F "TB_DATA=orange" localhost:8080/fruits`
- Request:  `curl --oauth2-bearer "<YOUR_JWT_HERE>" localhost:8080/fruits`
- Response: `apple`
- Request:  `curl --oauth2-bearer "<YOUR_JWT_HERE>" localhost:8080/fruits`
- Response: `orange`

---

## Future Plans

In the future, I may make the broker library API a bit more configurable (i.e. byte pre and post-processiong, etc.).
