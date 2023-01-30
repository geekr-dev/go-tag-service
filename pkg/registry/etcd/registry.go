package etcd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/geekr-dev/go-tag-service/pkg/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Registry is etcd registry.
type Registry struct {
	client  *clientv3.Client
	ctx     context.Context
	cancel  context.CancelFunc
	leaseID clientv3.LeaseID
}

func New() (*Registry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Registry{
		client: client,
		ctx:    ctx,
		cancel: cancelFunc,
	}, nil
}

func (r *Registry) Register(service string, data *registry.Service, expire int64) error {
	key := fmt.Sprintf("/etcdv3://geekr-dev/grpc/%s", service)
	val, _ := registry.Marshal(data)
	// 创建租约
	resp, err := r.client.Grant(r.ctx, expire)
	if err != nil {
		log.Printf("createLease failed,error %v \n", err)
		return err
	}
	r.leaseID = resp.ID
	// 绑定租约
	_, err = r.client.Put(r.ctx, key, val, clientv3.WithLease(r.leaseID))
	if err != nil {
		return err
	}
	// 续租(发送心跳)
	respChan, err := r.client.KeepAlive(r.ctx, r.leaseID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-respChan:
				// 续租成功
				continue
			case <-r.ctx.Done():
				// 取消租约，退出
				return
			}
		}
	}()
	return nil
}

// 获取服务地址
func (r *Registry) GetEndpoint(service string) (string, error) {
	key := fmt.Sprintf("/etcdv3://geekr-dev/grpc/%s", service)
	resp, err := r.client.Get(r.ctx, key)
	if err != nil {
		return "", err
	}
	endpoints := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		s, _ := registry.Unmarshal(kv.Value)
		if s.Name != service {
			continue
		}
		endpoints = append(endpoints, s.Endpoint)
	}

	// 实现简单负载均衡（random）
	rand.Seed(time.Now().Unix())
	i := rand.Intn(len(endpoints))
	return endpoints[i], nil
}

// 释放etcd客户端连接
func (r *Registry) Close() error {
	r.cancel()
	// 撤销租约
	r.client.Revoke(r.ctx, r.leaseID)
	return r.client.Close()
}
