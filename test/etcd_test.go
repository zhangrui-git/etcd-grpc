package main_test

import (
	"by/video/service/video"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestReg(t *testing.T) {
	var err error
	// 建立连接
	etcd, err := clientv3.New(
		clientv3.Config{
			Endpoints:   []string{"localhost:12379"},
			DialTimeout: time.Second * 3,
		},
	)
	if err != nil {
		t.Error(err)
	}
	defer func(etcd *clientv3.Client) {
		err := etcd.Close()
		if err != nil {
			t.Error(err)
		}
	}(etcd)

	ctx := context.Background()
	// 创建租约
	lease, _ := etcd.Grant(ctx, 10)

	em, err := endpoints.NewManager(etcd, "video")
	if err != nil {
		t.Error(err)
	}
	// 添加节点
	err = em.AddEndpoint(ctx, "video/127.0.0.1", endpoints.Endpoint{Addr: "127.0.0.1:82", Metadata: video.Video_ServiceDesc.Metadata}, clientv3.WithLease(lease.ID))
	if err != nil {
		t.Error(err)
	}

	// 租约续期
	for {
		select {
		case <-time.After(time.Second * 5):
			aliveOnce, keepErr := etcd.KeepAliveOnce(ctx, lease.ID)
			if keepErr != nil {
				return
			}
			fmt.Println(aliveOnce)
		case <-ctx.Done():
			return
		}
	}
}

func TestDis(t *testing.T) {
	// 创建连接
	etcd, etcdErr := clientv3.NewFromURL("http://127.0.0.1:12379")
	if etcdErr != nil {
		t.Error(etcdErr)
	}
	// 服务发现
	builder, resolverErr := resolver.NewBuilder(etcd)
	if resolverErr != nil {
		t.Error(resolverErr)
	}
	target := fmt.Sprintf("etcd:///%s", video.Video_ServiceDesc.ServiceName)
	conn, dialErr := grpc.Dial(
		target,
		grpc.WithResolvers(builder),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if dialErr != nil {
		t.Error(dialErr)
	}
	req := &video.PushRequest{Title: "banana", Comment: "banana"}

	for {
		select {
		case <-time.After(time.Second):
			grpcClient := video.NewVideoClient(conn)
			ctx := context.Background()
			response, callErr := grpcClient.Push(ctx, req)
			if callErr != nil {
				t.Error(callErr)
			}
			fmt.Println(response)
		}
	}
}
