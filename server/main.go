package main

import (
	"by/video/logic"
	"by/video/rpc"
	"by/video/service/video"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:80", "listen addr, ip:port")
	etcd := flag.String("etcd", "127.0.0.1:2379", "etcd addr, host:port")
	flag.Parse()

	svc := rpc.NewService(
		[]string{*etcd},
		rpc.WithServiceAddr(*listen),
		rpc.WithServiceDesc(&video.Video_ServiceDesc),
		rpc.WithTTL(3),
		rpc.WithKeepalive(2),
	)
	go svc.Register()

	grpcServer := grpc.NewServer()
	video.RegisterVideoServer(grpcServer, logic.VideoService{})

	listener, listenErr := net.Listen("tcp", *listen)
	if listenErr != nil {
		log.Fatal("监听失败", listenErr)
	} else {
		log.Println("监听成功")
	}
	serverErr := grpcServer.Serve(listener)
	if serverErr != nil {
		log.Fatal("启动失败", serverErr)
	}
}
