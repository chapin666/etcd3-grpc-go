package lib

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"

	etcd3 "github.com/coreos/etcd/clientv3"
)

var Prefix = "rpc"
var client etcd3.Client
var serviceKey string
var stopSignal = make(chan bool, 1)

func Register(name string, host string, port int, target string, interval time.Duration, ttl int) error {
	serviceValue := fmt.Sprintf("%s:%d", host, port)
	serviceKey = fmt.Sprintf("/%s/%s/%s", Prefix, name, serviceValue)
	var err error
	client, err := etcd3.New(etcd3.Config{
		Endpoints: strings.Split(target, ","),
	})

	if err != nil {
		return fmt.Errorf("grpclb: create etcd3 client failed: %v", err)
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			resp, _ := client.Grant(context.TODO(), int64(ttl))
			_, err := client.Get(context.Background(), serviceKey)
			if err != nil {
				if err == rpctypes.ErrKeyNotFound {
					if _, err := client.Put(context.TODO(), serviceKey, serviceValue, etcd3.WithLease(resp.ID)); err != nil {
						log.Printf("grpclb: set service '%s' with ttl to etcd3 failed: %s", name, err.Error())
					}
				} else {
					log.Printf("grpclb: service '%s' connect to etcd3 failed: %s", name, err.Error())
				}
			} else {
				if _, err := client.Put(context.Background(), serviceKey, serviceValue, etcd3.WithLease(resp.ID)); err != nil {
					log.Printf("grpclb: refresh service '%s' with ttl to etcd3 failed: %s", name, err.Error())
				}
			}

			select {
			case <-stopSignal:
				return
			case <-ticker.C:

			}
		}
	}()

	return nil
}

func UnRegister() error {
	stopSignal <- true
	stopSignal = make(chan bool, 1)
	var err error

	if _, err := client.Delete(context.Background(), serviceKey); err != nil {
		log.Printf("grpclb: deregisetr '%s' failed: %s", serviceKey, err.Error())
	} else {
		log.Printf("grpclb: deregister '%s' ok. ", serviceKey)
	}
	return err
}
