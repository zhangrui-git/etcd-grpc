package main_test

import (
	"by/video/service/video"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

func ConnInit() (client video.VideoClient, cond *sync.Cond) {
	conn, connErr := grpc.Dial(":83", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if connErr != nil {
		log.Fatal("客户端连接失败", connErr)
	}
	client = video.NewVideoClient(conn)
	cond = sync.NewCond(&sync.Mutex{})

	go func() {
		cond.L.Lock()
		cond.Wait()
		err := conn.Close()
		cond.L.Unlock()
		if err != nil {
			return
		}
	}()
	return
}

func TestInfo(t *testing.T) {
	grpcClient, cond := ConnInit()
	defer cond.Signal()
	ctx := context.Background()
	infoClient, callErr := grpcClient.Info(ctx)
	if callErr != nil {
		log.Fatal("接口调用失败", callErr)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		response, err := infoClient.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("客户端接收结束")
				return
			}
			log.Println("客户端接收失败", err) // 客户端接收失败rpc error: code = Unavailable desc = error reading from server: read tcp 127.0.0.1:58141->127.0.0.1:82: wsarecv: An existing connection was forcibly closed by the remote host.
		}
		fmt.Println(response)
	}()
	go func() {
		defer wg.Done()

		sendErr := infoClient.Send(&video.InfoRequest{Id: 1})
		if sendErr != nil {
			log.Println("客户端发送失败", sendErr) // 客户端发送失败 EOF
		}
		closeErr := infoClient.CloseSend()
		if closeErr != nil {
			log.Println("客户端关闭失败", sendErr)
			return
		}
		time.Sleep(time.Second * 2)
	}()
	wg.Wait()
}

func TestPush(t *testing.T) {
	grpcClient, cond := ConnInit()
	defer cond.Signal()
	ctx := context.Background()
	req := &video.PushRequest{Title: "banana", Comment: "banana"}
	response, err := grpcClient.Push(ctx, req)
	if err != nil {
		log.Println("客户端请求失败", err) // 客户端请求失败 rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp :82: connectex: No connection could be made because the target machine actively refused it.
		return
	}
	fmt.Println(response)
}
