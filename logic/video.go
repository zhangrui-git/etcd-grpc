package logic

import (
	"by/video/service/video"
	"context"
	"fmt"
	"io"
	"log"
)

type VideoService struct {
	video.UnimplementedVideoServer
}

// Info 双向流
func (s VideoService) Info(server video.Video_InfoServer) error {
	res := &video.InfoResponse{Id: 1, Title: "drink", Comment: "bear", Status: 0}
	for {
		request, recErr := server.Recv()
		if recErr != nil {
			if recErr == io.EOF {
				log.Println("服务端接收完毕", recErr)
			} else {
				log.Println("服务端接收失败", recErr) //  服务端接收失败rpc error: code = Canceled desc = context canceled
			}
			break
		}
		fmt.Println(request)

		sendErr := server.Send(res)
		if sendErr != nil {
			log.Println("服务端发送失败", sendErr)
			break
		}
	}
	return nil
}

// Push 同步请求
func (s VideoService) Push(ctx context.Context, request *video.PushRequest) (*video.InfoResponse, error) {
	fmt.Println(request)
	return &video.InfoResponse{Id: 8, Title: "milk", Comment: "milk", Status: 0}, nil
}
