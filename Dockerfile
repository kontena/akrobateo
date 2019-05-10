FROM golang:1.11 as builder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR  /go/src/github.com/kontena/akrobateo

# Add dependency graph and vendor it in
ADD Gopkg.* /go/src/github.com/kontena/akrobateo/
RUN dep ensure -v -vendor-only

# Add source and compile
ADD . /go/src/github.com/kontena/akrobateo/

ARG ARCH=amd64

RUN ./build.sh "linux/${ARCH}"

FROM scratch

ARG ARCH=amd64

COPY --from=builder /go/src/github.com/kontena/akrobateo/output/akrobateo_linux_${ARCH} /akrobateo

ENTRYPOINT ["/akrobateo"]
