FROM golang:1.12-alpine AS builder

RUN apk update && apk add --no-cache git

ENV  GO111MODULE=on
WORKDIR /root
COPY . .
RUN go generate ./... && \
    CGO_ENABLED=0 go build ./cmd/pass-policy-service

FROM alpine:3.9
ENV POLICY_FILE=docker.json
COPY --from=builder /root/pass-policy-service /root/scripts /root/policies /

RUN chmod 700 /entrypoint.sh

CMD /entrypoint.sh

