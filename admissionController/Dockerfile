# build stage

FROM golang:1.10-stretch AS build-env
RUN mkdir -p /go/src/github.com/wso2/validation
WORKDIR /go/src/github.com/wso2/validation
COPY  . .

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o validationwebhook

FROM scratch
COPY --from=build-env /go/src/github.com/wso2/validation/validationwebhook .
ENTRYPOINT ["/validationwebhook"]
