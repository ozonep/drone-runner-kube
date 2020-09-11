FROM golang:buster as buster
ENV GOPATH /go
COPY . ${GOPATH}/src/github.com/ozonep/drone-runner-kube
WORKDIR ${GOPATH}/src/github.com/ozonep/drone-runner-kube
RUN chmod +x ./build.sh
RUN ./build.sh

FROM stefanprodan/alpine-base:latest
EXPOSE 3000
ENV GODEBUG netdns=go
ENV DRONE_PLATFORM_OS linux
ENV DRONE_PLATFORM_ARCH amd64
# COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=buster /go/src/github.com/ozonep/drone-runner-kube/release/linux/amd64/drone-runner-kube /bin/
ENTRYPOINT ["/bin/drone-runner-kube"]