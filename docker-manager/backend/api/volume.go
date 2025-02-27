package api

import (
	"context"
	"dockerpanel/backend/pkg/docker"
	"net/http"

	"github.com/docker/docker/api/types/volume"
	"github.com/gin-gonic/gin"
)

func RegisterVolumeRoutes(r *gin.Engine) {
	group := r.Group("/api/volumes")
	{
		group.GET("", listVolumes)
		group.POST("", createVolume)
		group.DELETE("/:name", removeVolume)
	}
}

func listVolumes(c *gin.Context) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	volumes, err := cli.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, volumes)
}

func createVolume(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	vol, err := cli.VolumeCreate(context.Background(), volume.CreateOptions{
		Name: req.Name,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vol)
}

func removeVolume(c *gin.Context) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	name := c.Param("name")
	if err := cli.VolumeRemove(context.Background(), name, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "数据卷已删除"})
}