package consul

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
	"time"
)

func TestClient_RegisterService(t *testing.T) {
	go startCheckServer()
	client, err := New([]string{"127.0.0.1:8500"}, "http")
	if err != nil {
		panic(err)
	}
	deReg := make(chan bool, 1)
	err = client.RegisterService([]string{"test1", "test2"}, "http", "a", "10.32.64.113", 80, deReg)
	if err != nil {
		panic(err)
	}
	p, index, err := client.GetService("test1", "http", "a")
	fmt.Println(p, index, err)
	select {}
}

func startCheckServer() {
	engine := gin.New()
	server := &http.Server{
		Addr:           ":80",
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, map[string]string{
			"msg": "hello, world",
		})
	})
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
