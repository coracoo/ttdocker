package api

import (
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "time"
    "net/http"
    "sort"
    "strings"
	"bufio"
    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/filters"
	"github.com/gin-gonic/gin"
)

// ComposeProject 定义项目结构
type ComposeProject struct {
    Name       string    `json:"name"`
    Path       string    `json:"path"`
    Compose    string    `json:"compose"`
    AutoStart  bool      `json:"autoStart"`
    Containers int       `json:"containers"`
    Status     string    `json:"status"`
    CreateTime time.Time `json:"createTime"`
}

// RegisterComposeRoutes 注册路由
func RegisterComposeRoutes(r *gin.Engine) {
    group := r.Group("/api/compose")
    {
        group.GET("/list", listProjects)
        group.GET("/deploy/events", deployEvents)
        group.POST("/:name/start", startProject)
        group.POST("/:name/stop", stopProject)
        group.GET("/:name/status", getStackStatus)
        group.DELETE("/remove/:name", removeProject)  // 修改为匹配当前请求格式
    }
}

// startProject 启动项目
func startProject(c *gin.Context) {
    name := c.Param("name")
    projectDir := filepath.Join("data", "project", name)
    
    // 使用 docker compose up 命令启动项目
    cmd := exec.Command("docker", "compose", "up", "-d")
    cmd.Dir = projectDir
    
    if output, err := cmd.CombinedOutput(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("启动失败: %s\n%s", err.Error(), string(output)),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "项目已启动"})
}

// stopProject 停止项目
func stopProject(c *gin.Context) {
    name := c.Param("name")
    projectDir := filepath.Join("data", "project", name)
    
    // 使用 docker compose stop 命令停止项目
    cmd := exec.Command("docker", "compose", "stop")
    cmd.Dir = projectDir
    
    if output, err := cmd.CombinedOutput(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("停止失败: %s\n%s", err.Error(), string(output)),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "项目已停止"})
}


// listProjects 获取项目列表
func listProjects(c *gin.Context) {
    // 创建 Docker 客户端
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    // 获取所有带有 compose 标签的容器
    containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
        All: true,
        Filters: filters.NewArgs(
            filters.Arg("label", "com.docker.compose.project"),
        ),
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 用于存储项目信息的 map
    projects := make(map[string]*ComposeProject)

    // 遍历容器，按项目分组
    for _, container := range containers {
        projectName := container.Labels["com.docker.compose.project"]
        if projectName == "" {
            continue
        }

        if _, exists := projects[projectName]; !exists {
            projects[projectName] = &ComposeProject{
                Name:       projectName,
                Path:      filepath.Join("data", "project", projectName),
                Containers: 0,
                Status:    "已停止",
                CreateTime: time.Unix(container.Created, 0),
            }
        }

        // 更新容器数量
        projects[projectName].Containers++

        // 如果有任何容器在运行，则项目状态为运行中
        if container.State == "running" {
            projects[projectName].Status = "运行中"
        }
    }

    // 转换为数组
    result := make([]*ComposeProject, 0, len(projects))
    for _, project := range projects {
        // 尝试读取 compose 文件
        composePath := filepath.Join(project.Path, "docker-compose.yml")
        if data, err := os.ReadFile(composePath); err == nil {
            project.Compose = string(data)
        }
        result = append(result, project)
    }

    // 按创建时间倒序排序
    sort.Slice(result, func(i, j int) bool {
        return result[i].CreateTime.After(result[j].CreateTime)
    })

    c.JSON(http.StatusOK, result)
}
// deployEvents 处理部署事件
func deployEvents(c *gin.Context) {
    projectName := c.Query("name")
    compose := c.Query("compose")
    
    if projectName == "" || compose == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "项目名称和配置内容不能为空"})
        return
    }

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("Access-Control-Allow-Origin", "*")

    messageChan := make(chan map[string]interface{})
    doneChan := make(chan bool)

    go func() {
        defer close(messageChan)

          sendMessage := func(msgType, msg string) {
            select {
            case <-doneChan: // 检查是否已完成
                return
            default:
                messageChan <- map[string]interface{}{
                    "type":    msgType,
                    "message": msg,
                }
            }
        }

        projectDir := filepath.Join("data", "project", projectName)
        composePath := filepath.Join(projectDir, "docker-compose.yml")
		
        // 检查项目目录是否已存在
        if _, err := os.Stat(projectDir); err == nil {
            // 目录已存在，提示用户并终止部署
            sendMessage("error", fmt.Sprintf("项目 '%s' 已存在，如需重新部署请先删除现有项目", projectName))
            return
        } else if !os.IsNotExist(err) {
            // 其他错误
            sendMessage("error", "检查项目目录失败: "+err.Error())
            return
        }
		
        // 创建项目目录（如果不存在）
        if err := os.MkdirAll(projectDir, 0755); err != nil {
            sendMessage("error", "创建项目目录失败: "+err.Error())
            return
        }

        // 保存 compose 文件，使用从请求中获取的compose内容
        if err := os.WriteFile(composePath, []byte(compose), 0644); err != nil {
            sendMessage("error", "保存配置文件失败: "+err.Error())
            return
        }

        sendMessage("info", "正在启动服务...")

        // 使用 docker compose 命令
        cmd := exec.Command("docker", "compose", "up", "-d")
        cmd.Dir = projectDir
        
        // 获取命令的标准输出和错误输出管道
        stdout, err := cmd.StdoutPipe()
        if err != nil {
            sendMessage("error", "创建输出管道失败: "+err.Error())
            return
        }
        stderr, err := cmd.StderrPipe()
        if err != nil {
            sendMessage("error", "创建错误输出管道失败: "+err.Error())
            return
        }
        
        // 启动命令
        if err := cmd.Start(); err != nil {
            sendMessage("error", "启动命令失败: "+err.Error())
            return
        }
        
        // 创建扫描器读取输出
        scannerDone := make(chan bool)
        go func() {
            defer close(scannerDone)
            scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
            for scanner.Scan() {
                line := scanner.Text()
                // 根据输出内容判断类型
                msgType := "info"
                if strings.Contains(line, "error") || strings.Contains(line, "Error") {
                    msgType = "error"
                } else if strings.Contains(line, "Created") || strings.Contains(line, "Started") {
                    msgType = "success"
                }
				
                select {
                case <-doneChan: // 检查是否已完成
                    return
                default:
                    sendMessage(msgType, line)
                }
            }
        }()
        
        // 等待命令完成
        err = cmd.Wait()
        <-scannerDone // 等待扫描器完成

      if err != nil {
            sendMessage("error", "部署失败: "+err.Error())
            return
        }

        // 检查容器状态
        cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
        if err != nil {
            sendMessage("error", "Docker客户端初始化失败: "+err.Error())
            return
        }
        defer cli.Close()

        containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
            All: true,
            Filters: filters.NewArgs(
                filters.Arg("label", "com.docker.compose.project="+projectName),
            ),
        })
        if err != nil {
            sendMessage("error", "获取容器状态失败: "+err.Error())
            return
        }

        // 检查所有容器是否都在运行
        allRunning := true
        for _, container := range containers {
            if container.State != "running" {
                allRunning = false
                break
            }
        }

        if allRunning {
            sendMessage("success", "所有服务已成功启动")
        } else {
            sendMessage("warning", "部分服务可能未正常启动，请检查状态")
        }
    }()

    c.Stream(func(w io.Writer) bool {
        select {
        case msg, ok := <-messageChan:
            if !ok {
                close(doneChan) // 标记为已完成
                return false
            }
            c.SSEvent("message", msg)
            return true
        case <-time.After(30 * time.Second): // 添加超时处理
            close(doneChan) // 标记为已完成
            return false
        }
    })
}
// getStackStatus 获取堆栈状态
func getStackStatus(c *gin.Context) {
    name := c.Param("name")  // 修改这里，使用 name 而不是 stack
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cli.Close()

    containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
        All: true,
        Filters: filters.NewArgs(
            filters.Arg("label", "com.docker.compose.project=" + name),  // 修改这里
        ),
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, containers)
}
// removeStack 函数需要重命名为 removeProject 以保持一致性
// 并且需要使用 name 参数
func removeProject(c *gin.Context) {
    name := c.Param("name")  // 修改这里，使用 name 而不是 stack
    projectDir := filepath.Join("data", "project", name)
    
    // 使用 docker compose down 命令停止并删除容器
    cmd := exec.Command("docker", "compose", "down")
    cmd.Dir = projectDir
    
    if output, err := cmd.CombinedOutput(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("删除失败: %s\n%s", err.Error(), string(output)),
        })
        return
    }
    
    // 删除项目目录
    if err := os.RemoveAll(projectDir); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "删除项目目录失败: " + err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "项目已删除"})
}