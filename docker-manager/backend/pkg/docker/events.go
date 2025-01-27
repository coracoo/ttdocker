// backend/pkg/docker/events.go
package docker // 必须作为第一行

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// 新增广播函数定义
var BroadcastFunc func(types.Message)

func SetBroadcastHandler(fn func(types.Message)) {
	BroadcastFunc = fn
}

func WatchDockerEvents(cli *client.Client) {
	eventsChan, errChan := cli.Events(context.Background(), types.EventsOptions{})

	go func() { // 使用goroutine避免阻塞
		for {
			select {
			case event, ok := <-eventsChan:
				if !ok {
					return
				}
				if BroadcastFunc != nil {
					BroadcastFunc(event)
				}
			case err, ok := <-errChan:
				if !ok {
					return
				}
				log.Printf("Docker event error: %v", err)
			}
		}
	}()
}
