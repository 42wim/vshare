package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	template = `<html>
<h1>Welcome to vshare</h1>
<img src="/static/once.jpg"><br><br>
This secret can only be read <b>once</b>, if you get an error message about an <b>invalid token</b> this means that someone has possibly <b>intercepted</b> your secret. <br>
Please contact the person who gave you this URL<br><br>

Click <a href=#here#>here</a> if you are the recipient of the secret.<br><br>

You can also use curl to save the secret (for example on your server):
<pre>
curl -sSL #here# > #output#
</pre>
</html>
`
)

func handleInfo(c echo.Context) error {
	url := os.Getenv("VSHARE_URL")
	file := c.QueryParam("file")
	info := template
	if file != "" {
		info = strings.ReplaceAll(info, "#here#", url+"/token/"+c.Param("token")+"?file="+file)
		info = strings.ReplaceAll(info, "#output#", file)
	} else {
		info = strings.ReplaceAll(info, "#here#", url+"/token/"+c.Param("token"))
		info = strings.ReplaceAll(info, "#output#", "output.txt")
	}
	return c.HTML(http.StatusOK, info)
}

func handleToken(c echo.Context) error {
	token := c.Param("token")
	file := c.QueryParam("file")
	secret, encoded, err := unWrap(token)
	if err != nil {
		return c.String(http.StatusInternalServerError, "invalid token")
	}
	if encoded {
		data, _ := base64.StdEncoding.DecodeString(secret)
		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", "attachment", file))
		return c.Blob(http.StatusOK, http.DetectContentType(data), data)
	}
	return c.String(http.StatusOK, secret)
}

func startWeb(listen string) {
	assetHandler := http.FileServer(rice.MustFindBox("static").HTTPBox())
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.GET("/info/:token", handleInfo)
	e.GET("/token/:token", handleToken)
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	fmt.Printf("vault server: %s\n", os.Getenv("VAULT_ADDR"))
	fmt.Printf("vshare server URL: %s\n", os.Getenv("VSHARE_URL"))
	e.Logger.Fatal(e.Start(listen))
}
