FROM golang:rc

WORKDIR /usr/src/tinybroker

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/tinybroker

ENV TB_USER "user"
ENV TB_PASS "pass"
ENV TB_SECRET "mySecret"

CMD [ "tinybroker", "-v", "-a", ":8080" ]
