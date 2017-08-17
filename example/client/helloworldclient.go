package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	pb "coding.net/lycamplus-etcd-service-go/example/proto"
	discovery "coding.net/lycamplus-etcd-service-go/lib"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serv = flag.String("service", "hello_service", "service_name")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()

	r := discovery.NewResolver(*serv)
	b := grpc.RoundRobin(r)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, *reg, grpc.WithInsecure(), grpc.WithBalancer(b))
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1 * time.Second)
	for t := range ticker.C {
		client := pb.NewGreeterClient(conn)
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
		}
	}
}
