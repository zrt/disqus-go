package main

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/labstack/echo"
)

var (
	secretkey      = CONFIG_SecretKey
	publickey      = CONFIG_PublicKey
	shortname      = CONFIG_ShortName
	maxCachedLimit = 100000
	listPostsApi   = "https://disqus.com/api/3.0/threads/listPosts.json"
	listPostsLimit = "100"
	cached         map[string][]byte
	lock           sync.RWMutex
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

func main() {
	cached = make(map[string][]byte)
	e := echo.New()

	e.GET("/listPosts", handleListPosts)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.Logger.Fatal(e.Start("127.0.0.1:7001"))
}
