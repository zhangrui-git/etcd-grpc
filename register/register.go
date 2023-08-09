package register

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"log"
	"time"
)

type Service struct {
	etcdAddr    []string
	serviceAddr string
	serviceDesc grpc.ServiceDesc
	ttl         int64
	keepalive   int64
	client      *clientv3.Client
}

func NewService(etcdAddr []string, serviceAddr string, desc grpc.ServiceDesc, ttl int64, keepalive int64) *Service {
	return &Service{
		etcdAddr:    etcdAddr,
		serviceAddr: serviceAddr,
		serviceDesc: desc,
		ttl:         ttl,
		keepalive:   keepalive,
	}
}

func (s *Service) Register() {
	etcd := s.etcdClient()

	epm, epmErr := endpoints.NewManager(etcd, s.target())
	if epmErr != nil {
		log.Fatal(epmErr)
	}

	ctx := context.Background()

	// 创建租约
	lease, _ := etcd.Grant(ctx, s.ttl)

	// 添加节点
	ep := endpoints.Endpoint{
		Addr:     s.serviceAddr,
		Metadata: s.serviceDesc.Metadata,
	}
	addErr := epm.AddEndpoint(ctx, s.serverKey(), ep, clientv3.WithLease(lease.ID))
	if addErr != nil {
		log.Fatal(addErr)
	}

	// 租约续期
	for {
		select {
		case <-time.After(time.Second * time.Duration(s.keepalive)):
			keep, keepErr := etcd.KeepAliveOnce(ctx, lease.ID)
			if keepErr != nil {
				return
			}
			fmt.Println(keep)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) Unregister() {
	etcd := s.etcdClient()

	epm, epmErr := endpoints.NewManager(etcd, s.target())
	if epmErr != nil {
		log.Fatal(epmErr)
	}

	ctx := context.Background()

	delErr := epm.DeleteEndpoint(ctx, s.serverKey())
	if delErr != nil {
		log.Fatal(delErr)
	}
}

func (s *Service) etcdClient() *clientv3.Client {
	if s.client == nil {
		// 建立连接
		cli, etcdErr := clientv3.New(
			clientv3.Config{
				Endpoints:   s.etcdAddr,
				DialTimeout: time.Second * 3,
			},
		)
		if etcdErr != nil {
			log.Fatal("ETCD连接失败", etcdErr)
		}
		s.client = cli
	}
	return s.client
}

func (s *Service) target() string {
	return s.serviceDesc.ServiceName
}

func (s *Service) serverKey() string {
	return fmt.Sprintf("%s/%s", s.serviceDesc.ServiceName, s.serviceAddr)
}
