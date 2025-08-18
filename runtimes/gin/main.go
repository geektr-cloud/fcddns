package main

import (
	"github.com/geektr-cloud/fcddns/server"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	bs := &server.JwtServer{}
	bs.Init()

	r := gin.Default()
	r.GET("/ddns/v1", func(c *gin.Context) {
		req := server.FcRequest{
			Path:     c.Request.URL.Path,
			ClientIP: c.ClientIP(),
		}

		resp := bs.DDNS(c, req)

		c.String(resp.StatusCode, resp.Body)
	})

	r.Run()
}
