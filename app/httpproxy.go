package main

import (
	"context"
	"flag"
	"io"
	"time"

	"github.com/chromedp/chromedp"

	"bytes"
	"fmt"
	"sync"

	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
)

var actxt context.Context

type UserResource struct {
	// normally one would use DAO (data access object)
	users map[string]User
}

func (u UserResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.
		Path("/chrome").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	tags := []string{"users"}

	ws.Route(ws.GET("/").To(u.getHtml).
		// docs
		Doc("通过代理获取链接的html").
		Param(ws.QueryParameter("url", "link url").DataType("string").DefaultValue("cip.cc")).
		Param(ws.QueryParameter("last", "last").DataType("bool").DefaultValue("false")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]User{}).
		Returns(200, "OK", []User{}).
		DefaultReturns("OK", []User{}))

	ws.Route(ws.GET("/wx/").To(u.getWxLastHtml).
		// docs
		Doc("获取微信公众号的最新文章的html").
		Param(ws.QueryParameter("url", "link url").DataType("string").DefaultValue("cip.cc")).
		Param(ws.QueryParameter("last", "last").DataType("bool").DefaultValue("false")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]User{}).
		Returns(200, "OK", []User{}).
		DefaultReturns("OK", []User{}))

	ws.Route(ws.GET("/wxlist/").To(u.getWxListHtml).
		// docs
		Doc("获取微信公众号的文章列表的html").
		Param(ws.QueryParameter("wxid", "wxid").DataType("string").DefaultValue("maogeshijue")).
		Param(ws.QueryParameter("last", "last").DataType("bool").DefaultValue("false")).
		Param(ws.QueryParameter("time", "time").DataType("int").DefaultValue("3")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]User{}).
		Returns(200, "OK", []User{}).
		DefaultReturns("OK", []User{}))

	ws.Route(ws.PUT("").To(u.createUser).
		// docs
		Doc("create a user").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(User{})) // from the request

	ws.Route(ws.DELETE("/{user-id}").To(u.removeUser).
		// docs
		Doc("delete a user").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

	return ws
}

// GET http://localhost:8080/users
//
var getHtml_lasthtml = ""

func (u UserResource) getHtml(request *restful.Request, response *restful.Response) {
	var JsInsertUrl = `//创建一个节点
	var oCurUrl=document.createElement('a');
	//设置a的属性
	oCurUrl.href=window.location.href;
	oCurUrl.id='chrome-cur-url';
	oCurUrl.innerText="chrome-cur-url";
	//添加元素  将创建的节点添加到Id为d的div里
	document.body.appendChild(oCurUrl);	
	`
	//log.Printf(JsInsertUrl)
	log.Printf("getHtml")
	url := request.QueryParameter("url")
	log.Printf(url)

	if "true" == request.QueryParameter("last") {
		log.Printf("Use last html")
		if "" == getHtml_lasthtml {
			log.Printf("Last Html is empty.")
		} else {
			io.WriteString(response, getHtml_lasthtml)
			return
		}
	}

	ctxt, cancelCtxt := chromedp.NewContext(actxt) // create new tab
	defer cancelCtxt()                             // close tab afterwards

	var body string
	var res interface{}
	//log.Println(url)
	//log.Println("https://weixin.sogou.com/weixin?query=maogeshijue")
	if err := chromedp.Run(ctxt,
		//chromedp.Navigate("https://weixin.sogou.com/weixin?query=maogeshijue"),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Evaluate(JsInsertUrl, &res),
		chromedp.OuterHTML("html", &body),
	); err != nil {
		log.Fatalf("Failed getting body of %v: %v", url, err)
	}
	getHtml_lasthtml = body
	log.Println("Body of %v starts with:", url)
	log.Println(body[0:200])
	io.WriteString(response, body)
}

// GET http://localhost:8080/users/1
//
func (u UserResource) getWxLastHtml(request *restful.Request, response *restful.Response) {
	log.Printf("getWxLastHtml")
	url := request.QueryParameter("url")
	log.Printf(url)

	ctxt, cancelCtxt := chromedp.NewContext(actxt) // create new tab
	defer cancelCtxt()                             // close tab afterwards

	var body string
	log.Println(url)
	//log.Println("https://weixin.sogou.com/weixin?query=maogeshijue")
	if err := chromedp.Run(ctxt,
		//chromedp.Navigate("https://weixin.sogou.com/weixin?query=maogeshijue"),
		chromedp.Navigate(url),
		//log.Printf("wx"),
		//chromedp.Sleep(2 * time.Second),
		chromedp.WaitVisible(`#sogou_vr_11002301_box_0`, chromedp.ByID),
		chromedp.SetAttributeValue(`//*[@id="sogou_vr_11002301_box_0"]/dl[last()]/dd/a`, "target", "_parent", chromedp.NodeVisible),
		chromedp.Click(`//*[@id="sogou_vr_11002301_box_0"]/dl[last()]/dd/a`, chromedp.NodeVisible),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &body),
	); err != nil {
		log.Fatalf("Failed getting body of %v: %v", url, err)
	}

	log.Println("Body of starts with:", url)
	log.Println(body)
	//response.WriteEntity(body)
	//response.ResponseWriter(body)
	io.WriteString(response, body)

}

// GET http://localhost:8080/users/1
//

var getWxListHtml_lasthtml = ""

func (u UserResource) getWxListHtml(request *restful.Request, response *restful.Response) {
	log.Printf("getWxListHtml")
	wxid := request.QueryParameter("wxid")
	url := "https://weixin.sogou.com/weixin?type=2&query=" + wxid
	log.Printf(url)
	log.Printf("time:" + request.QueryParameter("time"))

	if "true" == request.QueryParameter("last") {
		log.Printf("Use last html")
		if "" == getWxListHtml_lasthtml {
			log.Printf("Last Html is empty.")
		} else {
			io.WriteString(response, getWxListHtml_lasthtml)
			return
		}
	}

	ctxt, cancelCtxt := chromedp.NewContext(actxt) // create new tab
	defer cancelCtxt()                             // close tab afterwards
        // create a timeout
	ctxt, cancelCtxt = context.WithTimeout(ctxt, 20*time.Second)
	defer cancelCtxt()

	var body string
	var body2,body3,body4,body5 string
	if err := chromedp.Run(ctxt,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`#tool_show`, chromedp.ByID),
		chromedp.Click(`tool_show`, chromedp.ByID),
		chromedp.OuterHTML("html", &body),
		chromedp.WaitVisible(`#search`, chromedp.ByID),
		chromedp.Click(`search`, chromedp.ByID),
		chromedp.Sleep(2 * time.Second),
		chromedp.OuterHTML("html", &body2),
		chromedp.SetValue(`//*[@id="tool"]/span[5]/div/form/span/input`, wxid),
		chromedp.WaitVisible(`#search_enter`, chromedp.ByID),
		chromedp.Click(`search_enter`, chromedp.ByID),
		chromedp.OuterHTML("html", &body3),
		chromedp.Sleep(2 * time.Second),
		chromedp.WaitVisible(`#time`, chromedp.ByID),
		chromedp.Click(`time`, chromedp.ByID),
		chromedp.OuterHTML("html", &body4),
		chromedp.Sleep(2 * time.Second),
		chromedp.WaitVisible(`#tool`, chromedp.ByID),
		chromedp.Sleep(2 * time.Second),
		chromedp.OuterHTML("html", &body5),
		chromedp.Click(`//*[@id="tool"]/span[1]/div/a[`+request.QueryParameter("time")+`]`, chromedp.NodeVisible),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &body),
	); err != nil {
		log.Printf("Failed getting body of %v: %v", url, err)
	}

	log.Println("Body of starts with:", url)
	log.Println(body[0:1000])
	log.Println(body2[0:1000])
	log.Println(body3[0:1000])
	log.Println(body4[0:1000])
	log.Println(body5[0:1000])
	getWxListHtml_lasthtml = body
	io.WriteString(response, body)

}

// PUT http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa Raspberry</Name></User>
//
func (u *UserResource) updateUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(&usr)
	if err == nil {
		u.users[usr.ID] = *usr
		response.WriteEntity(usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

// PUT http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa</Name></User>
//
func (u *UserResource) createUser(request *restful.Request, response *restful.Response) {
	usr := User{ID: request.PathParameter("user-id")}
	err := request.ReadEntity(&usr)
	if err == nil {
		u.users[usr.ID] = usr
		response.WriteHeaderAndEntity(http.StatusCreated, usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

// DELETE http://localhost:8080/users/1
//
func (u *UserResource) removeUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	delete(u.users, id)
}

func main() {
	var devToolWsUrl string
	flag.StringVar(&devToolWsUrl, "devtools-ws-url", "", "DevTools Websocket URL")
	flag.Parse()
	log.Printf(devToolWsUrl)

	var cancelActxt context.CancelFunc

	var bufMu sync.Mutex
	var buf bytes.Buffer
	fn := func(format string, a ...interface{}) {
		bufMu.Lock()
		fmt.Fprintf(&buf, format, a...)
		fmt.Fprintln(&buf)
		bufMu.Unlock()
	}

	ctx, cancel := chromedp.NewContext(context.Background(),
		chromedp.WithErrorf(fn),
		chromedp.WithLogf(fn),
		chromedp.WithDebugf(fn),
	)
	defer cancel()
	actxt, cancelActxt = chromedp.NewRemoteAllocator(ctx, devToolWsUrl)
	defer cancelActxt()

	u := UserResource{map[string]User{}}
	restful.DefaultContainer.Add(u.WebService())

	config := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json
	http.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("/Users/emicklei/Projects/swagger-ui/dist"))))

	// Optionally, you may need to enable CORS for the UI to work.
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CookiesAllowed: false,
		Container:      restful.DefaultContainer}
	restful.DefaultContainer.Filter(cors.Filter)

	log.Printf("Get the API using http://localhost:8080/apidocs.json")
	log.Printf("Open Swagger UI using http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "UserService",
			Description: "Resource for managing Users",
			Contact: &spec.ContactInfo{
				Name:  "john",
				Email: "john@doe.rp",
				URL:   "http://johndoe.org",
			},
			License: &spec.License{
				Name: "MIT",
				URL:  "http://mit.org",
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{spec.Tag{TagProps: spec.TagProps{
		Name:        "users",
		Description: "Managing users"}}}
}

// User is just a sample type
type User struct {
	ID   string `json:"id" description:"identifier of the user"`
	Name string `json:"name" description:"name of the user" default:"john"`
	Age  int    `json:"age" description:"age of the user" default:"21"`
}
