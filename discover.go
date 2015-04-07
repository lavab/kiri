package kiri

import (
	"errors"
	"log"
	"sync"
)

var (
	rr     map[string]int
	rrLock sync.Mutex

	ErrNoServices = errors.New("No services left")
)

type DiscoverFunc func(*Service) error

func (k *Kiri) Discover(name string, tags map[string]interface{}, cb DiscoverFunc) error {
	services, err := k.Query(name, tags)
	if err != nil {
		return err
	}

	rrLock.Lock()
	if l, ok := rr[name]; ok {
		rr[name] = l + 1

		if rr[name] >= len(services) {
			rr[name] = 0
		}
	} else {
		rr[name] = 0
	}
	start := rr[name]
	rrLock.Unlock()

	for i := start; ; {
		if l, ok := rr[name]; ok {
			rr[name] = l + 1

			if rr[name] >= len(services) {
				rr[name] = 0
			}
		} else {
			rr[name] = 0
		}

		err := cb(services[rr[name]])
		if err != nil {
			log.Print(err)

			err := k.Remove(name, services[rr[name]].Address)
			if err != nil {
				return err
			}
		}

		i++
		if i >= len(services) {
			i = 0
		}
		if i == start {
			break
		}
	}

	return ErrNoServices
}
