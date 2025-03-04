package api

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
	"strings"
    "net/http"
    "os"
    "sync"
    "github.com/docker/docker/api/types"
    "github.com/gorilla/websocket"
    "dockerpanel/backend/pkg/docker"
    "github.com/gin-gonic/gin"
)

type Terminal struct {
    conn      *websocket.Conn
    execID    string
    hijacked types.HijackedResponse
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
    bufferSize int
}

func NewTerminal(conn *websocket.Conn, execID string, hijacked types.HijackedResponse) *Terminal {
    ctx, cancel := context.WithCancel(context.Background())
    return &Terminal{
        conn:       conn,
        execID:     execID,
        hijacked:   hijacked,
        ctx:        ctx,
        cancel:     cancel,
        bufferSize: 4096,
    }
}

func (t *Terminal) Start() error {
    t.wg.Add(2)

    // 处理输入
    go func() {
        defer t.wg.Done()
        for {
            select {
            case <-t.ctx.Done():
                return
            default:
                messageType, message, err := t.conn.ReadMessage()
                if err != nil {
                    fmt.Printf("读取WebSocket消息错误: %v\n", err)
                    t.cancel()
                    return
                }

                // 只处理文本和二进制消息
                if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
                    continue
                }

                _, err = t.hijacked.Conn.Write(message)
                if err != nil {
                    fmt.Printf("写入容器错误: %v\n", err)
                    t.cancel()
                    return
                }
            }
        }
    }()

    // 处理输出
    go func() {
        defer t.wg.Done()
        for {
            select {
            case <-t.ctx.Done():
                return
            default:
                buf := make([]byte, t.bufferSize)
                nr, err := t.hijacked.Reader.Read(buf)
                if err != nil {
                    if err != io.EOF {
                        fmt.Printf("读取容器输出错误: %v\n", err)
                        t.cancel()
                    }
                    return
                }

                // 只发送实际读取的数据
                if nr > 0 {
                    err = t.conn.WriteMessage(websocket.BinaryMessage, buf[:nr])
                    if err != nil {
                        fmt.Printf("发送WebSocket消息错误: %v\n", err)
                        t.cancel()
                        return
                    }
                }
            }
        }
    }()

    t.wg.Wait()
    return nil
}

func (t *Terminal) Close() {
    t.cancel()
    t.hijacked.Close()
    t.conn.Close()
}

// 添加一个新的终端处理函数，使用Docker SDK直接执行命令
func containerExec(c *gin.Context) {
    containerId := c.Param("id")
    command := c.Query("cmd")
    
    if command == "" {
        command = "/bin/sh" // 默认命令
    }
    
    fmt.Printf("执行容器命令: %s, 容器ID: %s\n", command, containerId)
    
    cli, err := docker.NewDockerClient()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Docker客户端创建失败: %v", err)})
        return
    }
    defer cli.Close()
    
    // 检查容器是否存在并运行
    containerInfo, err := cli.ContainerInspect(context.Background(), containerId)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("容器不存在: %v", err)})
        return
    }
    
    if !containerInfo.State.Running {
        c.JSON(http.StatusBadRequest, gin.H{"error": "容器未运行，无法执行命令"})
        return
    }
    
    // 解析命令
    cmdParts := strings.Fields(command)
    
    // 容器执行命令的配置
    execConfig := types.ExecConfig{
        Cmd:          cmdParts,
        AttachStdout: true,
        AttachStderr: true,
        AttachStdin:  false, // 不需要输入
        Tty:          false, // 不使用TTY
    }
    
    // 创建容器执行命令
    execResp, err := cli.ContainerExecCreate(context.Background(), containerId, execConfig)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建exec命令失败: %v", err)})
        return
    }
    
    // 执行容器命令并获取输出
    resp, err := cli.ContainerExecAttach(context.Background(), execResp.ID, types.ExecStartCheck{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("附加到exec命令失败: %v", err)})
        return
    }
    defer resp.Close()
    
    // 读取所有输出
    output, err := io.ReadAll(resp.Reader)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取命令输出失败: %v", err)})
        return
    }
    
    // 返回命令输出
    c.JSON(http.StatusOK, gin.H{
        "output": string(output),
        "command": command,
        "container_id": containerId,
    })
}

// 添加一个交互式终端处理函数，使用Docker SDK的TTY模式
func containerInteractiveExec(c *gin.Context) {
    containerId := c.Param("id")
    
    // 升级HTTP连接为WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        fmt.Printf("WebSocket升级失败: %v\n", err)
        return
    }
    defer ws.Close()
    
    // 发送连接成功消息
    ws.WriteMessage(websocket.TextMessage, []byte("WebSocket连接成功，正在连接到容器...\n"))
    
    cli, err := docker.NewDockerClient()
    if err != nil {
        sendErrorMessage(ws, fmt.Sprintf("Docker客户端创建失败: %v", err))
        return
    }
    defer cli.Close()
    
    // 检查容器是否存在并运行
    containerInfo, err := cli.ContainerInspect(context.Background(), containerId)
    if err != nil {
        sendErrorMessage(ws, fmt.Sprintf("容器不存在: %v", err))
        return
    }
    
    if !containerInfo.State.Running {
        sendErrorMessage(ws, "容器未运行，无法连接终端")
        return
    }
    
    // 获取用户请求的命令，默认为/bin/sh
    messageType, p, err := ws.ReadMessage()
    cmdToUse := []string{"/bin/sh"}
    
    if err == nil && messageType == websocket.TextMessage {
        var msg struct {
            Type string `json:"type"`
            Data string `json:"data"`
        }
        
        if err := json.Unmarshal(p, &msg); err == nil && msg.Type == "command" && msg.Data != "" {
            cmdParts := strings.Fields(msg.Data)
            if len(cmdParts) > 0 {
                cmdToUse = cmdParts
            }
        }
    }
    
    // 容器执行命令的配置 - 使用TTY模式
    execConfig := types.ExecConfig{
        Cmd:          cmdToUse,
        AttachStdout: true,
        AttachStderr: true,
        AttachStdin:  true, // 需要输入
        Tty:          true, // 使用TTY
    }
    
    // 创建容器执行命令
    execResp, err := cli.ContainerExecCreate(context.Background(), containerId, execConfig)
    if err != nil {
        sendErrorMessage(ws, fmt.Sprintf("创建exec命令失败: %v", err))
        return
    }
    
    // 执行容器命令并获取输出
    resp, err := cli.ContainerExecAttach(context.Background(), execResp.ID, types.ExecStartCheck{
        Tty: true,
    })
    if err != nil {
        sendErrorMessage(ws, fmt.Sprintf("附加到exec命令失败: %v", err))
        return
    }
    defer resp.Close()
    
    // 使用互斥锁确保WebSocket写入的线程安全
    var wsWriteMu sync.Mutex
    
    // 创建一个完成通道，用于同步goroutine
    done := make(chan struct{})
    
    // 从容器输出读取并发送到WebSocket
    go func() {
        defer func() {
            fmt.Println("容器输出处理goroutine结束")
            close(done)
        }()
        
        buf := make([]byte, 4096)
        for {
            nr, err := resp.Reader.Read(buf)
            if err != nil {
                if err != io.EOF {
                    wsWriteMu.Lock()
                    fmt.Printf("读取容器输出错误: %v\n", err)
                    ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("读取容器输出错误: %v\n", err)))
                    wsWriteMu.Unlock()
                }
                break
            }
            
            if nr > 0 {
                wsWriteMu.Lock()
                err = ws.WriteMessage(websocket.BinaryMessage, buf[:nr])
                wsWriteMu.Unlock()
                if err != nil {
                    fmt.Printf("发送WebSocket消息错误: %v\n", err)
                    break
                }
            }
        }
    }()
    
    // 从WebSocket读取并写入容器输入
    go func() {
        defer func() {
            fmt.Println("WebSocket输入处理goroutine结束")
            // 通知另一个goroutine结束
            select {
            case <-done:
                // 已经关闭了
            default:
                close(done)
            }
        }()
        
        for {
            messageType, p, err := ws.ReadMessage()
            if err != nil {
                fmt.Printf("读取WebSocket消息错误: %v\n", err)
                break
            }
            
            if messageType == websocket.TextMessage {
                var msg struct {
                    Type string `json:"type"`
                    Data string `json:"data"`
                }
                
                if err := json.Unmarshal(p, &msg); err == nil {
                    fmt.Printf("收到WebSocket消息: type=%s, data长度=%d\n", msg.Type, len(msg.Data))
                    
                    switch msg.Type {
                    case "input":
                        // 写入到容器的标准输入
                        _, err = resp.Conn.Write([]byte(msg.Data))
                        if err != nil {
                            wsWriteMu.Lock()
                            fmt.Printf("写入容器输入错误: %v\n", err)
                            ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("写入容器输入错误: %v\n", err)))
                            wsWriteMu.Unlock()
                            break
                        }
                    case "resize":
                        // 处理终端大小调整
                        var size struct {
                            Rows uint `json:"rows"`
                            Cols uint `json:"cols"`
                        }
                        if err := json.Unmarshal([]byte(msg.Data), &size); err == nil {
                            fmt.Printf("调整终端大小: rows=%d, cols=%d\n", size.Rows, size.Cols)
                            err = cli.ContainerExecResize(context.Background(), execResp.ID, types.ResizeOptions{
                                Height: size.Rows,
                                Width:  size.Cols,
                            })
                            if err != nil {
                                wsWriteMu.Lock()
                                fmt.Printf("调整终端大小错误: %v\n", err)
                                ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("调整终端大小错误: %v\n", err)))
                                wsWriteMu.Unlock()
                            }
                        } else {
                            fmt.Printf("解析终端大小数据错误: %v\n", err)
                        }
                    }
                } else {
                    fmt.Printf("解析WebSocket消息错误: %v\n", err)
                }
            } else {
                fmt.Printf("收到非文本消息: type=%d, 长度=%d\n", messageType, len(p))
            }
        }
    }()
    
    // 等待任一goroutine完成
    <-done
    fmt.Println("终端会话结束")
}


// 定义WebSocket升级器
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许所有来源的WebSocket连接
    },
}


// 添加WebSocket终端处理函数
func containerTerminal(c *gin.Context) {
    containerId := c.Param("id")

    fmt.Printf("收到终端连接请求: %s\n", c.Param("id"))

    // 升级HTTP连接为WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        fmt.Printf("WebSocket升级失败: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("WebSocket升级失败: %v", err)})
        return
    }
    defer ws.Close()
     
    // 发送连接成功消息
    ws.WriteMessage(websocket.TextMessage, []byte("WebSocket连接成功，正在连接到容器...\n"))
	
    cli, err := docker.NewDockerClient()
    if err != nil {
        errMsg := fmt.Sprintf("Docker客户端创建失败: %v\n", err)
        fmt.Println(errMsg)
        ws.WriteMessage(websocket.TextMessage, []byte(errMsg))
        return
    }
    defer cli.Close()
	
    // 检查容器是否存在
    _, err = cli.ContainerInspect(context.Background(), containerId)
    if err != nil {
        errMsg := fmt.Sprintf("容器不存在或无法访问: %v\n", err)
        fmt.Println(errMsg)
        ws.WriteMessage(websocket.TextMessage, []byte(errMsg))
        return
    }    
	
    // 创建exec配置
    execConfig := types.ExecConfig{
        AttachStdin:  true,
        AttachStdout: true,
        AttachStderr: true,
        Tty:          true,
        Cmd:          []string{"/bin/sh"},
    }
    
    fmt.Printf("为容器 %s 创建exec实例\n", containerId)
    // 创建exec实例
    execResp, err := cli.ContainerExecCreate(context.Background(), containerId, execConfig)
    if err != nil {
        errMsg := fmt.Sprintf("创建exec实例失败: %v\n", err)
        fmt.Println(errMsg)
        ws.WriteMessage(websocket.TextMessage, []byte(errMsg))
        return
    }
    
    fmt.Printf("附加到exec实例 %s\n", execResp.ID)
    // 附加到exec实例
    hijacked, err := cli.ContainerExecAttach(context.Background(), execResp.ID, types.ExecStartCheck{
        Detach: false,
        Tty:    true,
    })
    if err != nil {
        errMsg := fmt.Sprintf("附加到exec实例失败: %v\n", err)
        fmt.Println(errMsg)
        ws.WriteMessage(websocket.TextMessage, []byte(errMsg))
        return
    }
    defer hijacked.Close()
    
    fmt.Println("成功附加到容器，开始数据传输")
    ws.WriteMessage(websocket.TextMessage, []byte("成功连接到容器终端\n"))
    
    // 处理WebSocket消息
    // 使用互斥锁确保WebSocket写入的线程安全
    var wsWriteMu sync.Mutex
    
    // 创建一个完成通道，用于同步goroutine
    done := make(chan struct{})
    
    // 从容器输出读取并发送到WebSocket
    go func() {
        defer func() {
            fmt.Println("容器输出处理goroutine结束")
            close(done)
        }()
        
        buf := make([]byte, 4096)
        for {
            nr, err := hijacked.Reader.Read(buf)
            if err != nil {
                if err != io.EOF {
                    wsWriteMu.Lock()
                    fmt.Printf("读取容器输出错误: %v\n", err)
                    ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("读取容器输出错误: %v\n", err)))
                    wsWriteMu.Unlock()
                }
                break
            }
            
            if nr > 0 {
                wsWriteMu.Lock()
                err = ws.WriteMessage(websocket.BinaryMessage, buf[:nr])
                wsWriteMu.Unlock()
                if err != nil {
                    fmt.Printf("发送WebSocket消息错误: %v\n", err)
                    break
                }
            }
        }
    }()
    
    // 从WebSocket读取并写入容器输入
    go func() {
        defer func() {
            fmt.Println("WebSocket输入处理goroutine结束")
            // 通知另一个goroutine结束
            select {
            case <-done:
                // 已经关闭了
            default:
                close(done)
            }
        }()
        
        for {
            messageType, p, err := ws.ReadMessage()
            if err != nil {
                fmt.Printf("读取WebSocket消息错误: %v\n", err)
                break
            }
            
            if messageType == websocket.TextMessage {
                var msg struct {
                    Type string `json:"type"`
                    Data string `json:"data"`
                }
                
                if err := json.Unmarshal(p, &msg); err == nil {
                    fmt.Printf("收到WebSocket消息: type=%s, data长度=%d\n", msg.Type, len(msg.Data))
                    
                    switch msg.Type {
                    case "input":
                        // 写入到容器的标准输入
                        _, err = hijacked.Conn.Write([]byte(msg.Data))
                        if err != nil {
                            wsWriteMu.Lock()
                            fmt.Printf("写入容器输入错误: %v\n", err)
                            ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("写入容器输入错误: %v\n", err)))
                            wsWriteMu.Unlock()
                            break
                        }
                    case "resize":
                        // 处理终端大小调整
                        var size struct {
                            Rows uint `json:"rows"`
                            Cols uint `json:"cols"`
                        }
                        if err := json.Unmarshal([]byte(msg.Data), &size); err == nil {
                            fmt.Printf("调整终端大小: rows=%d, cols=%d\n", size.Rows, size.Cols)
                            err = cli.ContainerExecResize(context.Background(), execResp.ID, types.ResizeOptions{
                                Height: size.Rows,
                                Width:  size.Cols,
                            })
                            if err != nil {
                                wsWriteMu.Lock()
                                fmt.Printf("调整终端大小错误: %v\n", err)
                                ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("调整终端大小错误: %v\n", err)))
                                wsWriteMu.Unlock()
                            }
                        } else {
                            fmt.Printf("解析终端大小数据错误: %v\n", err)
                        }
                    }
                } else {
                    fmt.Printf("解析WebSocket消息错误: %v\n", err)
                }
            } else {
                fmt.Printf("收到非文本消息: type=%d, 长度=%d\n", messageType, len(p))
            }
        }
    }()
    
    // 等待任一goroutine完成
    <-done
    fmt.Println("终端会话结束")
}

// 添加一个新的终端处理函数，使用不同的方法
func containerTerminalAlternative(c *gin.Context) {
    containerId := c.Param("id")
    fmt.Printf("收到终端连接请求(替代方法): %s\n", containerId)

    // 升级HTTP连接为WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        fmt.Printf("WebSocket升级失败: %v\n", err)
        return
    }
    defer ws.Close()
    
    cli, err := docker.NewDockerClient()
    if err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("错误: %v\n", err)))
        return
    }
    defer cli.Close()
    
    // 检查容器是否存在并运行
    containerInfo, err := cli.ContainerInspect(context.Background(), containerId)
    if err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("容器不存在: %v\n", err)))
        return
    }
    
    if !containerInfo.State.Running {
        ws.WriteMessage(websocket.TextMessage, []byte("容器未运行，无法连接终端\n"))
        return
    }
    
    // 获取用户请求的命令，默认为/bin/sh
    messageType, p, err := ws.ReadMessage()
    cmdToUse := []string{"/bin/sh"}
    
    if err == nil && messageType == websocket.TextMessage {
        var msg struct {
            Type string `json:"type"`
            Data string `json:"data"`
        }
        
        if err := json.Unmarshal(p, &msg); err == nil && msg.Type == "command" && msg.Data != "" {
            cmdParts := strings.Fields(msg.Data)
            if len(cmdParts) > 0 {
                cmdToUse = cmdParts
            }
        }
    }
    
    // 创建exec配置
    execConfig := types.ExecConfig{
        AttachStdin:  true,
        AttachStdout: true,
        AttachStderr: true,
        Tty:          true,
        Cmd:          cmdToUse,
    }
    
    // 创建exec实例
    execResp, err := cli.ContainerExecCreate(context.Background(), containerId, execConfig)
    if err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("创建exec实例失败: %v\n", err)))
        return
    }
    
    // 使用新的方法处理终端会话
    terminal := NewTerminal(ws, execResp.ID, types.HijackedResponse{})
    
    // 附加到exec实例
    resp, err := cli.ContainerExecAttach(context.Background(), execResp.ID, types.ExecStartCheck{
        Detach: false,
        Tty:    true,
    })
    if err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("附加到exec实例失败: %v\n", err)))
        return
    }
    defer resp.Close()
    
    // 更新terminal的hijacked字段
    terminal.hijacked = resp
    
    // 启动终端
    if err := terminal.Start(); err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("启动终端失败: %v\n", err)))
    }
}


// 发送错误消息到WebSocket
func sendErrorMessage(ws *websocket.Conn, message string) {
    ws.WriteMessage(websocket.TextMessage, []byte(message))
}

// 修改路由注册函数，添加替代方法的路由
func RegisterTerminalRoutes(r *gin.Engine) {
    // 注册WebSocket终端路由
    //r.GET("/api/containers/:id/terminal", containerTerminal)
    // 注册替代方法的路由
    //r.GET("/api/containers/:id/terminal2", containerTerminalAlternative)
	// 注册WebSocket终端路由
    r.GET("/api/containers/:id/terminal", containerInteractiveExec)
    // 注册非交互式命令执行路由
    r.GET("/api/containers/:id/exec", containerExec)
}