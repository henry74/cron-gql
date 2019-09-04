FROM golang:1.13.0-alpine3.10 as build

LABEL maintainer="scheduler service"

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY . .

RUN apk update && apk add git && \ 
git config --global http.sslVerify false && \
go get github.com/99designs/gqlgen/graphql && \
go get github.com/agnivade/levenshtein && \ 
go get github.com/dustin/go-humanize

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o scheduler server/server.go

FROM alpine:3.10
# Add non root user and certs
RUN addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build /go/src/app/scheduler  .

RUN chown -R app /home/app

USER app

RUN chmod +x -R /home/app

EXPOSE 8080

CMD ["/home/app/scheduler"]