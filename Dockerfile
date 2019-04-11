FROM golang:1.12-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /root
COPY . .
RUN CGO_ENABLED=0 go build ./cmd/policies 

FROM alpine:3.9
COPY --from=builder /root/policies /root/scripts /

RUN chmod 700 /entrypoint.sh

CMD /entrypoint.sh

