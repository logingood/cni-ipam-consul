// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/ip"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/murat1985/cni-ipam-consul/backend"
	"github.com/murat1985/cni-ipam-consul/config"
)

type IPAllocator struct {
	start net.IP
	end   net.IP
	conf  *config.IPAMConfig
	store backend.Store
}

func NewIPAllocator(conf *config.IPAMConfig, store backend.Store) (*IPAllocator, error) {
	var (
		start net.IP
		end   net.IP
		err   error
	)
	start, end, err = networkRange((*net.IPNet)(&conf.Subnet))
	if err != nil {
		return nil, err
	}

	// skip the .0 address
	start = ip.NextIP(start)

	if conf.RangeStart != nil {
		if err := validateRangeIP(conf.RangeStart, (*net.IPNet)(&conf.Subnet)); err != nil {
			return nil, err
		}
		start = conf.RangeStart
	}
	if conf.RangeEnd != nil {
		if err := validateRangeIP(conf.RangeEnd, (*net.IPNet)(&conf.Subnet)); err != nil {
			return nil, err
		}
		// RangeEnd is inclusive
		end = ip.NextIP(conf.RangeEnd)
	}

	return &IPAllocator{start, end, conf, store}, nil
}

func validateRangeIP(ip net.IP, ipnet *net.IPNet) error {
	if !ipnet.Contains(ip) {
		return fmt.Errorf("%s not in network: %s", ip, ipnet)
	}
	return nil
}

// Returns newly allocated IP along with its config
func (a *IPAllocator) Get(id string) (*types.IPConfig, error) {
	a.store.Lock()
	defer a.store.Unlock()

	gw := a.conf.Gateway
	if gw == nil {
		gw = ip.NextIP(a.conf.Subnet.IP)
	}

	var requestedIP net.IP
	if a.conf.Args != nil {
		requestedIP = a.conf.Args.IP
	}

	if requestedIP != nil {
		if gw != nil && gw.Equal(a.conf.Args.IP) {
			return nil, fmt.Errorf("requested IP must differ gateway IP")
		}

		subnet := net.IPNet{
			IP:   a.conf.Subnet.IP,
			Mask: a.conf.Subnet.Mask,
		}
		err := validateRangeIP(requestedIP, &subnet)
		if err != nil {
			return nil, err
		}

		reserved, err := a.store.Reserve(id, requestedIP)
		if err != nil {
			return nil, err
		}

		if reserved {
			return &types.IPConfig{
				IP:      net.IPNet{IP: requestedIP, Mask: a.conf.Subnet.Mask},
				Gateway: gw,
				Routes:  a.conf.Routes,
			}, nil
		}
		return nil, fmt.Errorf("requested IP address %q is not available in network: %s", requestedIP, a.conf.Name)
	}

	for cur := a.start; !cur.Equal(a.end); cur = ip.NextIP(cur) {
		// don't allocate gateway IP
		if gw != nil && cur.Equal(gw) {
			continue
		}

		reserved, err := a.store.Reserve(id, cur)
		if err != nil {
			return nil, err
		}
		if reserved {
			return &types.IPConfig{
				IP:      net.IPNet{IP: cur, Mask: a.conf.Subnet.Mask},
				Gateway: gw,
				Routes:  a.conf.Routes,
			}, nil
		}
	}
	return nil, fmt.Errorf("no IP addresses available in network: %s", a.conf.Name)
}

// Releases all IPs allocated for the container with given ID
func (a *IPAllocator) Release(id string) error {
	a.store.Lock()
	defer a.store.Unlock()

	return a.store.ReleaseByID(id)
}

func networkRange(ipnet *net.IPNet) (net.IP, net.IP, error) {
	if ipnet.IP == nil {
		return nil, nil, fmt.Errorf("missing field %q in IPAM configuration", "subnet")
	}
	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
		if ip == nil {
			return nil, nil, fmt.Errorf("IP not v4 nor v6")
		}
	}

	if len(ip) != len(ipnet.Mask) {
		return nil, nil, fmt.Errorf("IPNet IP and Mask version mismatch")
	}

	var end net.IP
	for i := 0; i < len(ip); i++ {
		end = append(end, ip[i]|^ipnet.Mask[i])
	}
	return ipnet.IP, end, nil
}
