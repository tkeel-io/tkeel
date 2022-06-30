############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM golang:1.17 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn,direct make release

FROM alpine:3.13

ENV PLUGIN_ID=keel

COPY --from=builder /src/dist/linux_amd64/release/keel /
CMD ["/keel"]