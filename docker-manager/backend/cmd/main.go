// cmd/main.go
package main

import (
    "dockerpanel/backend/api"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Configure CORS
    config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    r.Use(cors.New(config))

    // Register API routes
    api.RegisterContainerRoutes(r)
    api.RegisterImageRoutes(r)
    api.RegisterVolumeRoutes(r)
    api.RegisterNetworkRoutes(r)
    api.RegisterComposeRoutes(r)
	api.RegisterTerminalRoutes(r)

    // 使用特定前缀处理静态文件
    r.Static("/static", "./dist")
    
    // 添加根路由重定向到静态文件
    r.GET("/", func(c *gin.Context) {
        c.Redirect(301, "/static/index.html")
    })

    // 修改监听地址为 0.0.0.0:8080，允许从任何 IP 访问
    r.Run("0.0.0.0:8080")
}
