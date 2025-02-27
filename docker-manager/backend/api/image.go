package api

import (
    "bytes"
    "context"
    "dockerpanel/backend/pkg/docker"
    "net/http"
    "strings"

    "github.com/docker/docker/api/types"
    "github.com/gin-gonic/gin"
)

func RegisterImageRoutes(r *gin.Engine) {
    group := r.Group("/api/images")
    {
        group.GET("", listImages)
        group.DELETE("/:id", removeImage)
        group.POST("/pull", pullImage)
        group.GET("/proxy", getDockerProxy)      // 添加获取代理配置
        group.POST("/proxy", updateDockerProxy)   // 添加更新代理配置
    }
}

// Docker代理配置结构
type DockerConfig struct {
    Proxies map[string]string `json:"proxies"`  // HTTP/HTTPS 代理
    Mirrors []string          `json:"mirrors"`   // 镜像加速器
}

func getDockerProxy(c *gin.Context) {
    config, err := docker.ReadDaemonConfig()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置失败: " + err.Error()})
        return
    }
    c.JSON(http.StatusOK, config)
}

func updateDockerProxy(c *gin.Context) {
    var reqConfig DockerConfig
    if err := c.ShouldBindJSON(&reqConfig); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置格式: " + err.Error()})
        return
    }

    config := &docker.DockerConfig{
        Proxies: reqConfig.Proxies,
        Mirrors: reqConfig.Mirrors,
    }

    if err := docker.UpdateDaemonConfig(config); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配置失败: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Docker配置已更新，请重启 Docker 服务以生效"})
}

func listImages(c *gin.Context) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, images)
}

func removeImage(c *gin.Context) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cli.Close()

	id := c.Param("id")
	_, err = cli.ImageRemove(context.Background(), id, types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "镜像已删除"})
}

func pullImage(c *gin.Context) {
    var req struct {
        Image  string `json:"name" binding:"required"`
        Mirror string `json:"mirror"`
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

    // 如果指定了镜像加速器，临时修改镜像地址
    imageName := req.Image
    if req.Mirror != "" {
        // 解析原始镜像名称
        parts := strings.Split(req.Image, "/")
        if len(parts) > 1 {
            // 使用镜像加速器地址替换原始地址
            imageName = req.Mirror + "/" + parts[len(parts)-1]
        }
    }

    reader, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer reader.Close()

    // 读取拉取进度
    buf := new(bytes.Buffer)
    buf.ReadFrom(reader)
    c.JSON(http.StatusOK, gin.H{"message": "镜像拉取成功"})
}