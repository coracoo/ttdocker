// cmd/main.go
package main

import (
    "dockerpanel/backend/api"
    "dockerpanel/backend/pkg/database"
    "log"
    "path/filepath"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    // 设置数据目录
    dataDir := "data"
    log.Printf("数据目录: %s", dataDir)

    // 初始化数据库
    log.Printf("正在初始化数据库...")
    dbPath := filepath.Join(dataDir, "data.db") // 指定数据库文件路径
    if err := database.InitDB(dbPath); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    defer database.Close()

    log.Println("数据库初始化成功")
    defer database.Close()

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
    api.RegisterImageRegistryRoutes(r)
	api.RegisterSystemRoutes(r)
	//api.RegisterTerminalRoutes(r)
    // 使用特定前缀处理静态文件
    r.Static("/static", "./dist")
    
    // 添加根路由重定向到静态文件
    r.GET("/", func(c *gin.Context) {
        c.Redirect(301, "/static/index.html")
    })

    r.Run(":8080")
}
