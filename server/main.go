package main

import (
	"by/video/logic"
	"by/video/register"
	"by/video/service/video"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	host := flag.String("host", "127.0.0.1", "listen ip")
	port := flag.String("port", "80", "listen port")
	etcd := flag.String("etcd", "127.0.0.1:2379", "etcd addr, ip:port")
	flag.Parse()
	serverAddr := fmt.Sprintf("%s:%s", *host, *port)
	etcdAddr := fmt.Sprintf("%s", *etcd)
	fmt.Println(serverAddr)
	fmt.Println(etcdAddr)

	svc := register.NewService([]string{etcdAddr}, serverAddr, video.Video_ServiceDesc, 3, 2)
	go svc.Register()

	grpcServer := grpc.NewServer()
	video.RegisterVideoServer(grpcServer, logic.VideoService{})

	listener, listenErr := net.Listen("tcp", serverAddr)
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
