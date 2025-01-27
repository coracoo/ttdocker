package api

import (
	"context"
	"dockerpanel/backend/pkg/docker"
	"net/http"
	"path/filepath"

	"github.com/docker/docker/api/types/filters"
	"github.com/gin-gonic/gin"
)

func RegisterComposeRoutes(r *gin.Engine) {
	group := r.Group("/api/compose")
	{
		group.POST("/deploy", deployCompose)
		group.GET("/status/:stack", getStackStatus)
		group.DELETE("/remove/:stack", removeStack)
	}
}

func deployCompose(c *gin.Context) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	composeFile := filepath.Join("deployments", "komga.yaml")
	if err := cli.DeployCompose(c.Request.Context(), composeFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "部署失败",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Komga部署成功"})
}

// 部署状态查询
func getStackStatus(c *gin.Context) {
	stack := c.Param("stack")
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	// 示例：获取容器状态
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "com.docker.compose.project=" + stack,
		}),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, containers)
}

// 移除堆栈
func removeStack(c *gin.Context) {
	stack := c.Param("stack")
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	// 示例：删除相关容器
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "com.docker.compose.project=" + stack,
		}),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, container := range containers {
		if err := cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "堆栈已移除"})
}
