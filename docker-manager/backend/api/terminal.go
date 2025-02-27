package api

import (
    "context"
    "fmt"
    "io"
    "net/http"
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
    // 添加缓冲区大小配置
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
        bufferSize: 4096, // 增加缓冲区大小
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

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func RegisterTerminalRoutes(r *gin.Engine) {
    r.GET("/api/containers/:id/exec", execContainer)
}

func execContainer(c *gin.Context) {
    fmt.Printf("收到终端连接请求: %s\n", c.Param("id"))

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

    execConfig := types.ExecConfig{
        AttachStdin:  true,
        AttachStdout: true,
        AttachStderr: true,
        Tty:          true,
        Cmd:          []string{"/bin/sh"},
    }

    execResp, err := cli.ContainerExecCreate(context.Background(), c.Param("id"), execConfig)
    if err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("错误: %v\n", err)))
        return
    }

    hijacked, err := cli.ContainerExecAttach(context.Background(), execResp.ID, types.ExecStartCheck{
        Detach: false,
        Tty:    true,
    })
    if err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("错误: %v\n", err)))
        return
    }
    defer hijacked.Close()

    term := NewTerminal(ws, execResp.ID, hijacked)
    defer term.Close()

    if err := term.Start(); err != nil {
        ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("错误: %v\n", err)))
    }
}