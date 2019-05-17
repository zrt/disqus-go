package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/labstack/echo"
)

var (
	secretkey      = CONFIG_SecretKey
	publickey      = CONFIG_PublicKey
	disquskey      = "E8Uh5l5fHZ6gD8U3KycjAIAk46f68Zw7C6eW8WSjZvCLXebZ7p0r1yrYDrLilk2F"
	shortname      = CONFIG_ShortName
	maxCachedLimit = 100000
	listPostsApi   = "https://disqus.com/api/3.0/threads/listPosts.json"
	listPostsLimit = "100"
	createPostApi  = "https://disqus.com/api/3.0/posts/create.json"
	listThreadApi  = "https://disqus.com/api/3.0/threads/list.json"
	approvePostApi = "https://disqus.com/api/3.0/posts/approve.json"
	cached         map[string][]byte
	lock           sync.RWMutex
	defaultMail    = "bot@yeah.moe"
	defaultName    = "Guest"
)

func updatelink(link string) []byte {
	resp, err := http.Get(listPostsApi + "?api_key=" + publickey + "&forum=" + shortname + "&limit=" + listPostsLimit + "&thread=link:" + link)
	if err != nil {
		return []byte("{\"result\": \"failed to access to disqus api\"}")
	}
	defer resp.Body.Close()
	encodedJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte("{\"result\": \"failed to read body\"}")
	}

	lock.Lock()
	if len(cached) > maxCachedLimit {
		cached = make(map[string][]byte)
	}
	cached[link] = encodedJSON
	lock.Unlock()
	return encodedJSON
}

// /listPosts
func handleListPosts(c echo.Context) error {
	link := c.QueryParam("link")
	lock.RLock()
	result, ok := cached[link]
	lock.RUnlock()
	if ok {
		go updatelink(link)
		return c.JSONBlob(http.StatusOK, result)
	} else {
		result = updatelink(link)
		return c.JSONBlob(http.StatusOK, result)
	}
}

// /listThreads
func handleListThreads(c echo.Context) error {
	link := c.QueryParam("link")
	if link == "" {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"please provide link\"}"))
	}
	resp, err := http.Get(listThreadApi + "?api_key=" + publickey + "&forum=" + shortname + "&thread=link:" + link)
	if err != nil {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"failed to access to disqus api\"}"))
	}
	defer resp.Body.Close()
	encodedJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"failed to read body\"}"))
	}
	return c.JSONBlob(http.StatusOK, encodedJSON)
}

// /create
func handleCreatePost(c echo.Context) error {
	thread := c.FormValue("thread")
	message := c.FormValue("message")
	name := c.FormValue("name")
	if thread == "" {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"please provide thread\"}"))
	}
	if message == "" {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"please provide message\"}"))
	}
	if name == "" {
		name = defaultName
	}
	resp, err := http.PostForm(createPostApi,
		url.Values{"thread": {thread}, "message": {message}, "author_name": {name}, "author_email": {defaultMail}, "api_key": {disquskey}})
	if err != nil {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"failed to access to disqus api\"}"))
	}
	defer resp.Body.Close()
	encodedJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSONBlob(http.StatusOK, []byte("{\"result\": \"failed to read body\"}"))
	}
	return c.JSONBlob(http.StatusOK, encodedJSON)
}

func main() {
	cached = make(map[string][]byte)
	e := echo.New()

	e.GET("/listPosts", handleListPosts)
	e.GET("/listThreads", handleListThreads)
	e.POST("/create", handleCreatePost)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.Logger.Fatal(e.Start("127.0.0.1:7001"))
}
