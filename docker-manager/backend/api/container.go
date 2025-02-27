// backend/api/container.go
package api // 必须声明包名

import (
    "context"
    "dockerpanel/backend/pkg/docker"
    "net/http"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"  // 添加这个导入
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
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	id := c.Param("id")
	if err := cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "容器已启动"})
}

func stopContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    // 使用 container.StopOptions
    if err := cli.ContainerStop(context.Background(), id, container.StopOptions{
        Timeout: nil, // 使用默认超时时间
    }); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已停止"})
}

// 只保留这一个 removeContainer 实现
func removeContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    err = cli.ContainerStop(context.Background(), id, container.StopOptions{
        Timeout: nil, // 使用默认超时时间
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    err = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已删除"})
}

func createContainer(c *gin.Context) {
    // 实现创建逻辑...
}
