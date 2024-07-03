FROM golang:1.21 AS builder
RUN apt update && apt install make git
WORKDIR /callisto
COPY . ./
RUN cat go.mod | grep -i wasm
RUN go mod tidy && go mod vendor
RUN make build

FROM alpine:latest
WORKDIR /callisto
COPY --from=builder /callisto/build/callisto /usr/bin/callisto
CMD [ "callisto" ]