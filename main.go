package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	tokens            []string
	sessionCookieName = "sessionId"
)

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie(sessionCookieName)
		if err != nil {
			ctx.Request.Method = "GET"
			ctx.Redirect(http.StatusSeeOther, "/auth")
			return
		}

		for _, token := range tokens {
			if cookie == token {
				ctx.Next()
				return
			}
		}

		ctx.Request.Method = "GET"
		ctx.Redirect(http.StatusSeeOther, "/auth")
		ctx.Next()
		return
	}
}

func main() {
	router := gin.Default()
	router.Use()

	router.LoadHTMLFiles("auth/index.html")

	router.GET("/auth", func(ctx *gin.Context) {
		cookie, err := ctx.Cookie(sessionCookieName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{})
			return
		}

		for _, token := range tokens {
			if cookie == token {
				ctx.Redirect(http.StatusSeeOther, "/")
				return
			}
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{})
		return

	})
	router.POST("/auth", postAuthEndpoint)

	authorized := router.Group("/", Auth())
	{
		authorized.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"Message": "Welcome"})
		})
	}

	router.Run(":8080")
}

func getAuthDataFromEnv() map[string]string {
	envData := map[string]string{}

	file, err := os.Open(".env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		readData := strings.Split(scanner.Text(), "=")

		envData[readData[0]] = readData[1]
	}

	return envData
}

func postAuthEndpoint(ctx *gin.Context) {
	login := ctx.PostForm("login")
	password := ctx.PostForm("password")

	dataFromEnv := getAuthDataFromEnv()

	for envLogin, envPassword := range dataFromEnv {
		if (login == envLogin) && (password == envPassword) {
			token := base64.StdEncoding.EncodeToString([]byte(login + ":" + password))
			tokens = append(tokens, token)
			ctx.SetCookie(sessionCookieName, token, 60*60, "/", "localhost", false, true)
			ctx.Request.Method = "GET"
			ctx.Redirect(http.StatusSeeOther, "/")
			return
		}
	}

	ctx.Request.Method = "GET"
	ctx.Redirect(http.StatusSeeOther, "/auth")
	return
}
