package main

import (
	"context"
	"flag"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"

	"github.com/sirupsen/logrus"
	pb "github.com/wwcd/grpc-lb/cmd/helloworld"
	grpclb "github.com/wwcd/grpc-lb/etcdv3"
)

var (
	svc = flag.String("service", "hello_service", "service name")
	reg = flag.String("reg", "http://localhost:2379", "register etcd address")
)

func main() {
	flag.Parse()

	// 注册gprc resolver Builder接口
	// type Builder interface {
	//	Build(target Target, cc ClientConn, opts BuildOption) (Resolver, error)
	//	Scheme() string
	// }
	r := grpclb.NewResolver(*reg, *svc)
	resolver.Register(r)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	// The gRPC client library will use the specified scheme to pick the right resolver plugin and pass it the fully qualified name string.
	// r.Scheme很重要 etcd_v3
	target := r.Scheme() + "://authority/" + *svc
	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())

	// etcdv3_resolver://authority/hello_service
	// key :  /etcdv3_resolver/hello_service/localhost:50001
	logrus.Infof(target)

	cancel()
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	for t := range ticker.C {
		client := pb.NewGreeterClient(conn)
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			logrus.Infof("%v: Reply is %s\n", t, resp.Message)
		}
	}
}
