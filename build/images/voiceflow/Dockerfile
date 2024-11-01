# Telepace base image: https://github.com/openim-sigs/openim-base-image

# Set go mod installation source and proxy

ARG GOARCH
ARG GOOS

FROM golang:1.22 AS builder

ARG GO111MODULE=on
ARG GOPROXY=https://goproxy.io,direct

WORKDIR /telepace/app

ENV GO111MODULE=$GO111MODULE
ENV GOPROXY=$GOPROXY

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build BINS=voiceflow

RUN cp /telepace/app/_output/bin/platforms/$(go env GOOS)/$(go env GOARCH)/voiceflow /usr/bin/voiceflow


FROM ghcr.io/telepace/ai_feedback_production_ubuntu:latest

WORKDIR ${TELEPACE_APP}

COPY --from=builder /usr/bin/voiceflow ./bin/voiceflow

ENTRYPOINT ["./bin/voiceflow"]