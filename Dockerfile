FROM golang:1.10.3-alpine AS build-env
ENV PRJ_DIR $GOPATH/src/github.com/cstoku/scheduling-scaler
RUN apk add --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
ADD . $PRJ_DIR/
WORKDIR $PRJ_DIR
RUN dep ensure
RUN go build -o /tmp/scheduling-scaler

FROM alpine
RUN apk add --no-cache tzdata
COPY --from=build-env /tmp/scheduling-scaler /usr/local/bin
ENTRYPOINT ["/usr/local/bin/scheduling-scaler"]
