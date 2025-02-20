# Use the same architecture for the builder and the final image
FROM golang:1.23.5-bullseye AS builder

RUN apt update && apt install make git -y
WORKDIR /callisto

COPY . ./
RUN go mod tidy && go mod download
RUN make build

FROM debian:bullseye
WORKDIR /root
RUN apt-get update && apt-get install ca-certificates -y

COPY --from=builder /go/pkg/mod/github.com/!cosm!wasm/wasmvm/v2@v2.2.1/internal/api/libwasmvm.* /root/
COPY --from=builder /callisto/build/callisto /root/callisto

# Set LD_LIBRARY_PATH without referencing undefined variable
ENV LD_LIBRARY_PATH=/root
CMD ["./callisto"]
