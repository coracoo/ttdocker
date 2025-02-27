// cmd/main.go
package main

import (
    "dockerpanel/backend/api"
    "dockerpanel/backend/pkg/docker"
    "github.com/docker/docker/api/types/events"  // 添加这个导入
    "github.com/gin-contrib/cors"                // 添加这个导入
    "github.com/gin-gonic/gin"
)

// WebSocket 广播处理
func handleDockerEvents(message events.Message) {
    // 处理 Docker 事件
}

func main() {
    r := gin.Default()

    // 添加静态文件服务
    r.Static("/", "./dist")  // 添加这行，指向前端构建文件目录

    // 注册所有API路由
    api.RegisterContainerRoutes(r)
    api.RegisterImageRoutes(r)
    api.RegisterVolumeRoutes(r)
    api.RegisterNetworkRoutes(r)
    api.RegisterComposeRoutes(r)

    // 设置 Docker 事件处理器
    docker.SetBroadcastHandler(handleDockerEvents)

    // 配置 CORS
    config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    r.Use(cors.New(config))

    // 启动服务器
    r.Run(":8080")
}
