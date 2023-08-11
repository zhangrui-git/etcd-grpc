package rpc

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"log"
	"time"
)

func (s *Service) Register() {
	s.etcdClient()
	s.etcdManager()

	ctx := context.Background()

	// 创建租约
	lease, _ := s.client.Grant(ctx, s.ttl)

	// 添加节点
	ep := endpoints.Endpoint{
		Addr:     s.serviceAddr,
		Metadata: s.serviceDesc.Metadata,
	}
	err := s.manager.AddEndpoint(ctx, s.etcdKey(), ep, clientv3.WithLease(lease.ID))
	if err != nil {
		log.Fatal("服务添加失败", err)
	}

	// 租约续期
	for {
		select {
		case <-time.After(time.Second * time.Duration(s.keepalive)):
			keep, keepErr := s.client.KeepAliveOnce(ctx, lease.ID)
			if keepErr != nil {
				log.Println("租约续期失败")
				return
			}
			fmt.Println(keep)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) Unregister() {
	s.etcdClient()
	s.etcdManager()

	ctx := context.Background()

	err := s.manager.DeleteEndpoint(ctx, s.etcdKey())
	if err != nil {
		log.Fatal("服务移除失败", err)
	}
}

func (s *Service) etcdManager() {
	if s.manager == nil {
		manager, err := endpoints.NewManager(s.client, s.target())
		if err != nil {
			log.Fatal("manager创建失败", err)
		}
		s.manager = manager
	}
}
