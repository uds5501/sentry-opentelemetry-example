FROM golang:1.16-alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
EXPOSE 8088
RUN go build
CMD ["/build/main"]