package kiri

import (
	"log"
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
		Etcd:     etcd.NewClient(addresses),
		Services: map[string]*Service{},
		Stores:   []*Store{},
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
	go func() {
		if err := store.Start(); err != nil {
			log.Print(err)
		}
	}()
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

func (k *Kiri) Query(name string, tags map[string]interface{}) ([]*Service, error) {
	result := []*Service{}

	// Fill the lookup table with data from stores
	lookup := map[string]int{}
	for _, store := range k.Stores {
		store.Lock()
		if services, ok := store.RemoteServices[name]; ok {
			for _, service := range services {
				if id, ok := lookup[service.Address]; ok {
					if service.Tags != nil {
						if result[id].Tags == nil {
							result[id].Tags = service.Tags
						} else {
							for key, value := range service.Tags {
								if _, ok := result[id].Tags[key]; !ok {
									result[id].Tags[key] = value
								}
							}
						}
					}
				} else {
					lookup[service.Address] = len(result)
					result = append(result, service)
				}
			}
		}
		store.Unlock()
	}

	// Filter by tags
	if tags != nil {
		filtered := []*Service{}

		for _, service := range result {
			incorrect := false

			for key, v1 := range tags {
				if v2, ok := service.Tags[key]; ok {
					if v1 != v2 {
						incorrect = true
					}
				} else {
					incorrect = true
				}
			}

			if !incorrect {
				filtered = append(filtered, service)
			}
		}

		result = filtered
	}

	return result, nil
}
