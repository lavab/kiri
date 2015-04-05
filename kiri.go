package kiri

import (
	"sync"

	"github.com/coreos/go-etcd/etcd"
)

type Kiri struct {
	Etcd     *etcd.Client
	Stores   []*Store
	Services map[string]*Service
	sync.Mutex
}

func New(addresses []string) *Kiri {
	return &Kiri{
		Etcd: etcd.NewClient(addresses),
	}
}

func (k *Kiri) Store(format Format, path string) {
	store := &Store{
		Format: format,
		Path:   path,
		Kiri:   k,
		Update: make(chan struct{}),
	}
	k.Stores = append(k.Stores, store)
	go store.Start()
}

func (k *Kiri) Register(name string, address string, tags map[string]interface{}) {
	service := &Service{
		Name:    name,
		Address: address,
		Tags:    tags,
	}

	k.Lock()
	k.Services[name] = service
	for _, store := range k.Stores {
		store.Update <- struct{}{}
	}
	k.Unlock()
}
