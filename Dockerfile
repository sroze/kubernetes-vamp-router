FROM golang:latest

RUN wget https://github.com/koofr/go-pin/releases/download/v1.10/go-pin.sh -O /usr/bin/go-pin && \
    chmod +x /usr/bin/go-pin

RUN mkdir -p /go/src/github.com/sroze/kubernetes-vamp-router
ADD . /go/src/github.com/sroze/kubernetes-vamp-router/

WORKDIR /go/src/github.com/sroze/kubernetes-vamp-router

RUN cat versions | go-pin reset \
    && go get \
    && go build -o main . \
    && mv /go/bin/kubernetes-vamp-router /kubernetes-vamp-router \
    && rm -rf /go

CMD ["/kubernetes-vamp-router"]
