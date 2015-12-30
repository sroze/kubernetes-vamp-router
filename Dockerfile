FROM golang:latest

RUN wget https://github.com/koofr/go-pin/releases/download/v1.10/go-pin.sh -O /usr/bin/go-pin && \
    chmod +x /usr/bin/go-pin

# Workspace
RUN mkdir -p /go/src/github.com/sroze/kubernetes-vamp-router
WORKDIR /go/src/github.com/sroze/kubernetes-vamp-router

# Dependencies
ADD versions /go/src/github.com/sroze/kubernetes-vamp-router/versions
RUN cat versions | go-pin reset

# Application
ADD . /go/src/github.com/sroze/kubernetes-vamp-router/
RUN go get
RUN go install
RUN cd cmd/k8svamprouter \
    && go build -o main . \
    && mv ./main /kubernetes-vamp-router \
    && rm -rf /go

CMD ["/kubernetes-vamp-router"]
