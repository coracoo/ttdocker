// cmd/main.go
package main

import (
	"context"
	"dockerpanel/backend/api"
	"dockerpanel/backend/pkg/docker" // 模块路径+包路径
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func init() {
	docker.SetBroadcastHandler(func(e types.Message) {
		// 示例：转换为JSON格式
		data, _ := json.Marshal(e)
		fmt.Println("Docker Event:", string(data))
	})
}

func main() {
	r := gin.Default()

	// 注册路由
	api.RegisterContainerRoutes(r)
	api.RegisterComposeRoutes(r)

	// 初始化Docker客户端
	cli, err := docker.NewDockerClient()
	if err != nil {
		panic(fmt.Sprintf("Docker连接失败: %v", err))
	}
	defer cli.Close()

	// 启动事件监听
	go docker.WatchDockerEvents(cli)

	version, err := cli.ServerVersion(context.Background())
	if err != nil {
		panic(fmt.Sprintf("版本获取失败: %v", err))
	}

	fmt.Printf("成功连接Docker引擎\n版本: %s\nAPI版本: %s\n",
		version.Version,
		version.APIVersion,
	)

	r.Run(":8080")
}
