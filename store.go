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
	Format         Format
	Path           string
	Kiri           *Kiri
	Update         chan struct{}
	RemoteServices map[string][]*Service
	sync.Mutex
}

func (s *Store) DecodeKey(name string, input string) ([]*Service, error) {
	result := []*Service{}

	if s.Format == Default {
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			return nil, err
		}

		for _, service := range result {
			service.Name = name
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

	return result, nil
}

func (s *Store) EncodeKey(input []*Service) ([]byte, error) {
	result := []byte{}

	if s.Format == Default {
		var err error
		result, err = json.Marshal(input)
		if err != nil {
			return nil, err
		}
	} else if s.Format == Puro {
		for _, service := range input {
			result = append(result, []byte(service.Address+"\n")...)
		}

		if len(result) > 0 {
			// Trim last newline
			result = result[:len(result)-2]
		}
	}

	return result, nil
}

func (s *Store) Reload() error {
	s.Lock()
	defer s.Unlock()

	log.Print("Reloading the service discovery")

	s.RemoteServices = map[string][]*Service{}

	resp, err := s.Kiri.Etcd.Get(s.Path, true, true)
	if err != nil {
		log.Print(err)
		return err
	}

	foundNodes := map[string]struct{}{}

	for _, node := range resp.Node.Nodes {
		name := node.Key[len(s.Path)+1:]

		services, err := s.DecodeKey(name, node.Value)
		if err != nil {
			log.Print(err)
			return err
		}

		if local, ok := s.Kiri.Services[name]; ok {
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

			if needChanges {
				data, err := s.EncodeKey(services)
				if err != nil {
					log.Print(err)
					return err
				}

				log.Printf("Setting %s - %s", node.Key, string(data))

				_, err = s.Kiri.Etcd.Set(node.Key, string(data), 60*60*48)
				if err != nil {
					log.Print(err)
					return err
				}

				log.Printf("Updated the service discovery for key %s", node.Key)
			}
		}

		s.RemoteServices[name] = services
		foundNodes[name] = struct{}{}
	}

	// Add missing
	for _, service := range s.Kiri.Services {
		if _, ok := foundNodes[service.Name]; !ok {
			data, err := s.EncodeKey([]*Service{
				service,
			})
			if err != nil {
				return err
			}

			_, err = s.Kiri.Etcd.Set(s.Path+"/"+service.Name, string(data), 60*60*48)
			if err != nil {
				log.Print(err)
				return err
			}

			log.Printf("Added service discovery for key %s", s.Path+"/"+service.Name)
		}
	}

	return nil
}

func (s *Store) Start() error {
	if err := s.Reload(); err != nil {
		return err
	}

	s.Kiri.QueryLock.Unlock()

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

	return nil
}
