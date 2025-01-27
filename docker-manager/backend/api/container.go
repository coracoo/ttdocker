// backend/api/container.go
package api // 必须声明包名

import (
	"context"
	"dockerpanel/backend/pkg/docker"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
)

// 容器列表
func ListContainers(c *gin.Context) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, containers)
}

// 路由注册需导出的函数
func RegisterContainerRoutes(r *gin.Engine) {
	group := r.Group("/api/containers")
	{
		group.GET("", ListContainers)
		group.POST("/:id/start", startContainer)
		group.POST("/:id/stop", stopContainer)
		group.POST("/create", createContainer)
		group.DELETE("/:id", removeContainer)
	}
}

// 其他处理函数实现
func startContainer(c *gin.Context) {
	//id := c.Param("id")
	// 实现启动逻辑...
}

func stopContainer(c *gin.Context) {
	//id := c.Param("id")
	// 实现停止逻辑...
}

func createContainer(c *gin.Context) {
	//id := c.Param("id")
	// 实现创建逻辑...
}

func removeContainer(c *gin.Context) {
	//id := c.Param("id")
	// 实现移除逻辑...
}
