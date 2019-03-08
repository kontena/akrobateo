FROM golang:1.11 as builder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR  /go/src/github.com/kontena/service-lb-operator

# Add dependency graph and vendor it in
ADD Gopkg.* /go/src/github.com/kontena/service-lb-operator/
RUN dep ensure -v -vendor-only

# Add source and compile
ADD . /go/src/github.com/kontena/service-lb-operator/

RUN ./build.sh

FROM scratch

ARG ARCH=amd64

COPY --from=builder /go/src/github.com/kontena/service-lb-operator/output/service-lb-operator_linux_${ARCH} /service-lb-operator

ENTRYPOINT ["/service-lb-operator"]
