FROM golang:1.19-bullseye AS builder
WORKDIR /go/src/github.com/forbole/bdjuno
COPY . ./
RUN go mod download
RUN make build
RUN ldd build/bdjuno > /deps.txt
RUN echo $(ldd build/bdjuno | grep libwasmvm.so | awk '{ print $3 }')
RUN cat /deps.txt

FROM debian:bullseye
WORKDIR /root
RUN apt-get update && apt-get install ca-certificates -y
COPY --from=builder /deps.txt /root/deps.txt
COPY --from=builder /go/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.4.1/internal/api/libwasmvm.x86_64.so /root
COPY --from=builder /go/src/github.com/forbole/bdjuno/build/bdjuno /root/bdjuno
ENV LD_LIBRARY_PATH=/root
CMD [ "bdjuno" ]

