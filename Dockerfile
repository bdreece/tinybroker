FROM golang:rc

WORKDIR /usr/src/tinybroker

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/tinybroker

ENV TB_USER "YOUR_USERNAME"
ENV TB_PASS "YOUR_PASSWORD"
ENV TB_SECRET "YOUR_SECRET"

EXPOSE 8080

CMD [ "tinybroker", "-VV" ]
