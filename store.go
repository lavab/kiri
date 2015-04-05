package kiri

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/coreos/go-etcd/etcd"
)

type Store struct {
	Format Format
	Path   string
	Kiri   *Kiri
	Update chan struct{}
	sync.Mutex
}

func (s *Store) DecodeKey(name string, input string) ([]*Service, error) {
	result := []*Service{}

	if s.Format == Default {
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			return err
		}
	} else if s.Format == Puro {
		lines := strings.Split(input, "\n")
		for _, line := range lines {
			result = append(result, &Service{
				Name:    name,
				Address: line,
			})
		}
	}

	return result
}

func (s *Store) Reload() error {
	s.Lock()
	defer s.Unlock()

	resp, err := s.Kiri.Etcd.Get(s.Path, true, true)
	if err != nil {
		return err
	}

	for _, node := range resp.Node.Nodes {
		name := node.Key[len(s.Path)+1:]

		if local, ok := s.Kiri.Services[name]; ok {
			services, err := s.DecodeKey(name, node.Value)
			if err != nil {
				return err
			}

			needChanges := false
			foundLocal := false

			for _, service := range services {
				if local.Address == service.Address {
					foundLocal = true

					if !reflect.DeepEqual(local.Tags, service.Tags) {
						needChanges = true
						service.Tags = local.Tags
					}
				}
			}

			if !foundLocal {
				services = append(services, local)
				needChanges = true
			}

			// TODO: Encode

			// TODO: Set the key
			// s.Kiri.Etcd.Set(key, value, ttl)
		}
	}
}

func (s *Store) Start() error {
	if err := s.Reload(); err != nil {
		return err
	}

	receiver := make(chan *etcd.Response)
	stop := make(chan bool)

	go func() {
		select {
		case <-receiver:
			if err := s.Reload(); err != nil {
				log.Print(err)
			}
		case <-s.Update:
			if err := s.Reload(); err != nil {
				log.Print(err)
			}
		}
	}()

	_, err := s.Kiri.Etcd.Watch(s.Path, 0, true, receiver, stop)
	if err != nil {
		return err
	}
}
