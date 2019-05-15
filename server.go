package main

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)


var (
	secretkey = CONFIG_SecretKey
	publickey = CONFIG_PublicKey
	shortname = CONFIG_ShortName
	listPostsApi = "https://disqus.com/api/3.0/threads/listPosts.json"
	listPostsLimit = "100"
)

// /listPosts
func handleListPosts(c echo.Context) error {
	link := c.QueryParam("link")
	println(link)
	resp, err := http.Get(listPostsApi + "?api_key="+publickey+"&forum="+shortname+"&limit="+listPostsLimit+"&thread=link:"+link)
    if err != nil {
        return c.JSON(http.StatusBadGateway, map[string]string{
			"result":     "failed to access to disqus api",
		})
    }
    defer resp.Body.Close()
    encodedJSON, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return c.JSON(http.StatusBadGateway, map[string]string{
			"result":     "failed to read body",
		})
    }
	return c.JSONBlob(http.StatusOK, encodedJSON)
}


func main() {
	e := echo.New()

	e.GET("/listPosts", handleListPosts)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})


	e.Logger.Fatal(e.Start("127.0.0.1:7001"))
}
