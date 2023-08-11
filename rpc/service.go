package rpc

import (
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
	manager     endpoints.Manager
}

type Option interface {
	Apply(*Service)
}

type funcOption struct {
	f func(*Service)
}

func (fo *funcOption) Apply(s *Service) {
	fo.f(s)
}

func WithServiceAddr(listen string) Option {
	return &funcOption{f: func(s *Service) {
		s.serviceAddr = listen
	}}
}

func WithServiceDesc(desc grpc.ServiceDesc) Option {
	return &funcOption{f: func(s *Service) {
		s.serviceDesc = desc
	}}
}

func WithTTL(duration int64) Option {
	return &funcOption{f: func(s *Service) {
		s.ttl = duration
	}}
}

func WithKeepalive(duration int64) Option {
	return &funcOption{f: func(s *Service) {
		s.keepalive = duration
	}}
}

func NewService(etcdAddr []string, opts ...Option) *Service {
	s := &Service{
		etcdAddr: etcdAddr,
	}

	for _, opt := range opts {
		opt.Apply(s)
	}

	return s
}

func (s *Service) etcdClient() {
	if s.client == nil {
		// 建立连接
		client, err := clientv3.New(
			clientv3.Config{
				Endpoints:   s.etcdAddr,
				DialTimeout: time.Second * 3,
			},
		)
		if err != nil {
			log.Fatal("etcd连接失败", err)
		}
		s.client = client
	}
}

func (s *Service) target() string {
	return s.serviceDesc.ServiceName
}

func (s *Service) etcdKey() string {
	return fmt.Sprintf("%s/%s", s.serviceDesc.ServiceName, s.serviceAddr)
}
