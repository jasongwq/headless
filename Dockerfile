FROM golang:1.13

WORKDIR /go/src/app
RUN go get -u -d github.com/chromedp/chromedp
RUN go get -u -d github.com/astaxie/beego
RUN go get -u -d github.com/beego/bee
RUN go get -u -d github.com/go-openapi/spec
RUN go get -u -d github.com/emicklei/go-restful
RUN go get -u -d github.com/emicklei/go-restful-openapi
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app.sh"]
