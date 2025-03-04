// backend/api/container.go
package api // 必须声明包名

import (
    "context"
    "dockerpanel/backend/pkg/docker"
    "fmt"
    "net/http"
    "regexp"
    "strings"
    "time"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/gin-gonic/gin"
)

// 路由注册需导出的函数
func RegisterContainerRoutes(r *gin.Engine) {
    group := r.Group("/api/containers")
    {
        group.GET("", ListContainers)
        group.POST("/:id/start", startContainer)
        group.POST("/:id/stop", stopContainer)
        group.POST("/:id/restart", restartContainer)
        group.POST("/:id/pause", pauseContainer)
        group.POST("/:id/unpause", unpauseContainer)
        group.DELETE("/:id", removeContainer)
		group.GET("/:id/logs", getContainerLogs)
		group.GET("/:id/terminal", containerTerminal)
    }
}

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

    // 获取每个容器的详细信息
    var containersWithDetails []gin.H
    for _, container := range containers {
        inspect, err := cli.ContainerInspect(context.Background(), container.ID)
        if err != nil {
            continue
        }

        // 处理端口映射，添加 IP 地址
        formattedPorts := make([]gin.H, 0)
        for _, port := range container.Ports {
            portInfo := gin.H{
                "PrivatePort": port.PrivatePort,
                "Type":       port.Type,
            }
            
            if port.PublicPort != 0 {
                // 添加 IP 信息
                hostIP := "0.0.0.0"
                if port.IP != "" {
                    hostIP = port.IP
                }
                portInfo["PublicPort"] = port.PublicPort
                portInfo["IP"] = hostIP
            }
            
            formattedPorts = append(formattedPorts, portInfo)
        }

        // 计算运行时间
        var runningTime string
        if container.State == "running" {
            startTime, err := time.Parse(time.RFC3339, inspect.State.StartedAt)
            if err != nil {
                runningTime = "时间解析错误"
            } else {
                runningTime = time.Since(startTime).Round(time.Second).String()
            }
        } else {
            runningTime = "未运行"
        }

        containerInfo := gin.H{
            "Id":              container.ID,
            "Names":          container.Names,
            "Image":          container.Image,
            "State":          container.State,
            "Status":         container.Status,
            "Created":        container.Created,
            "Ports":          formattedPorts,
            "NetworkSettings": inspect.NetworkSettings,  // 使用 inspect 中的网络设置
            "HostConfig":     inspect.HostConfig,       // 添加 HostConfig
            "RunningTime":    runningTime,
        }
        containersWithDetails = append(containersWithDetails, containerInfo)
    }

    c.JSON(http.StatusOK, containersWithDetails)
}


// 重启容器
func restartContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    if err := cli.ContainerRestart(context.Background(), id, container.StopOptions{}); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已重启"})
}

// 暂停容器
func pauseContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    if err := cli.ContainerPause(context.Background(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已暂停"})
}

// 恢复容器
func unpauseContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    if err := cli.ContainerUnpause(context.Background(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已恢复"})
}

// 启动容器
func startContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接到 Docker: " + err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    // 先检查容器是否存在
    _, err = cli.ContainerInspect(context.Background(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "容器不存在: " + err.Error()})
        return
    }

    // 检查容器当前状态
    containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "获取容器列表失败: " + err.Error()})
        return
    }

    // 查找目标容器
    var targetContainer types.Container
    found := false
    for _, container := range containers {
        if container.ID == id {
            targetContainer = container
            found = true
            break
        }
    }

    if !found {
        c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定容器"})
        return
    }

    // 检查容器状态
    if targetContainer.State == "running" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "容器已经在运行中"})
        return
    }

    // 尝试启动容器
    err = cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
    if err != nil {
        errMsg := err.Error()
        switch {
        case strings.Contains(errMsg, "bind: address already in use"):
            // 提取端口信息，匹配格式为 0.0.0.0:端口号 的模式
            portRegex := regexp.MustCompile(`0.0.0.0:(\d+)`)
            matches := portRegex.FindStringSubmatch(errMsg)
            if len(matches) > 1 {
                c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("端口冲突，%s，请检查端口", matches[1])})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "端口冲突，请检查端口配置"})
            }
        case strings.Contains(errMsg, "no such file or directory"):
            // 提取路径信息
            pathRegex := regexp.MustCompile(`path\s+([^\s]+)\s+`)
            matches := pathRegex.FindStringSubmatch(errMsg)
            if len(matches) > 1 {
                c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("路径不存在，请检查宿主机路径%s", matches[1])})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "路径不存在，请检查宿主机路径配置"})
            }
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "启动容器失败: " + errMsg})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已启动"})
}

// 停止容器
func stopContainer(c *gin.Context) {
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    id := c.Param("id")
    // 先检查容器是否存在
    _, err = cli.ContainerInspect(context.Background(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "容器不存在"})
        return
    }

    // 尝试停止容器
    err = cli.ContainerStop(context.Background(), id, container.StopOptions{
        Timeout: nil,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "容器已停止"})
}

// 删除容器
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

// 新建容器
func createContainer(c *gin.Context) {
    // 实现创建逻辑...
}
