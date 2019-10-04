FROM jasongwq/golang:v2

WORKDIR /go/src/app
COPY app .
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["/go/src/app/app.sh"]
