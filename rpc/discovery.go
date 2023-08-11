package rpc

import (
	"fmt"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func (s *Service) Discovery() *grpc.ClientConn {
	s.etcdClient()

	builder, resolverErr := resolver.NewBuilder(s.client)
	if resolverErr != nil {
		log.Fatal(resolverErr)
	}

	conn, dialErr := grpc.Dial(
		fmt.Sprintf("etcd:///%s", s.target()),
		grpc.WithResolvers(builder),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if dialErr != nil {
		log.Fatal(dialErr)
	}

	return conn
}
