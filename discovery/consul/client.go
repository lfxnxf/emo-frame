package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	log "github.com/lfxnxf/emo-frame/logging"
	"path"
	"strconv"
	"strings"
	"time"
)

// Client provides a wrapper around the consulkv client
type Client struct {
	client *api.Client
}

var bLog *log.Logger

// New returns a new client to Consul for the given address
func New(nodes []string, scheme string) (*Client, error) {

	conf := api.DefaultConfig()

	conf.Scheme = scheme

	if len(nodes) > 0 {
		conf.Address = nodes[0]
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func (c *Client) SetLogger(l *log.Logger) {
	bLog = l
}

func (c *Client) GetKeyValue(key string) (string, uint64, error) {

	key = strings.TrimPrefix(key, "/")
	pairs, _, err := c.client.KV().Get(key, nil)

	if err != nil {
		return "", 0, err
	}

	if pairs == nil {
		return "", 0, nil
	}
	return string(pairs.Value), pairs.ModifyIndex, nil
}

// GetValues queries Consul for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		key := key
		key = strings.TrimPrefix(key, "/")
		pairs, _, err := c.client.KV().List(key, nil)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[path.Join("/", p.Key)] = string(p.Value)
		}
	}
	return vars, nil
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

// WatchPrefix watch prefix keys
func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	respChan := make(chan watchResponse)
	go func() {
		opts := api.QueryOptions{
			WaitIndex: waitIndex,
		}
		_, meta, err := c.client.KV().List(prefix, &opts)
		if err != nil {
			respChan <- watchResponse{waitIndex, err}
			return
		}
		respChan <- watchResponse{meta.LastIndex, err}
	}()

	select {
	case <-stopChan:
		return waitIndex, nil
	case r := <-respChan:
		return r.waitIndex, r.err
	}
}

func (c *Client) registerService(service *api.AgentServiceRegistration, dereg chan bool) error {
	var serviceID string
	registered := func() bool {
		if serviceID == "" {
			existService, _, _ := c.client.Catalog().Service(service.Name, "", nil)
			for _, s := range existService {
				s := s
				if s.ServiceAddress == service.Address && s.ServicePort == service.Port {
					serviceID = s.ID
					return true
				}
			}
			return false
		}
		services, err := c.client.Agent().Services()
		if err != nil {

			if bLog != nil {
				bLog.Errorf("consul: Cannot get service list. %s", err)
			} else {
				log.GenLogf("consul: Cannot get service list. %s", err)
			}
			return false
		}
		return services[serviceID] != nil
	}

	register := func() error {
		if err := c.client.Agent().ServiceRegister(service); err != nil {

			if bLog != nil {
				bLog.Errorf("consul: Cannot register %s service in consul. %s", service.Name, err)
			} else {
				log.GenLogf("consul: Cannot register %s service in consul. %s", service.Name, err)
			}
			return err
		}
		serviceID = service.ID

		if bLog != nil {
			bLog.Infof("consul: register service %q in consul, serviceId %q", service.Name, serviceID)
		} else {
			log.GenLogf("consul: register service %q in consul, serviceId %q", service.Name, serviceID)
		}
		return nil
	}

	deRegister := func() error {
		log.GenLogf("consul: Deregistering service %s,%s", service.Name, serviceID)
		if bLog != nil {
			bLog.Infof("consul: Deregistering service %s,%s", service.Name, serviceID)
		} else {
			log.GenLogf("consul: Deregistering service %s,%s", service.Name, serviceID)
		}
		return c.client.Agent().ServiceDeregister(serviceID)
	}

	err := register()
	if err == nil {
		go func() {
			for {
				select {
				case <-dereg:
					_ = deRegister()
					return
				case <-time.After(10 * time.Second):
					if !registered() {
						_ = register()
					}
				}
			}
		}()
	}
	return err
}

func (c *Client) comRegisterService(target, proto string, tags []string, address string, port int, deReg chan bool) error {

	var check *api.AgentServiceCheck
	check = &api.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d", address, port),
		Interval:                       "5s",
		Timeout:                        "200ms",
		DeregisterCriticalServiceAfter: "60s",
	}
	service := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s:%d", target, address, port),
		Name:    target,
		Port:    port,
		Address: address,
		Tags:    tags,
		Check:   check,
	}
	err := c.registerService(service, deReg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) makeTarget(target, proto string) string {
	nTarget := target
	if proto == "http" {
		nTarget += "-http"
	} else {
		nTarget += "-" + proto
	}
	return nTarget
}

func (c *Client) RegisterMultiTagService(targets []string, proto string, tags []string, address string, port int, deReg chan bool) error {

	for _, target := range targets {

		ntarget := c.makeTarget(target, proto)

		err := c.comRegisterService(ntarget, proto, tags, address, port, deReg)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterService target service
func (c *Client) RegisterService(targets []string, proto, tag, address string, port int, dereg chan bool) error {
	for _, target := range targets {
		nTarget := c.makeTarget(target, proto)
		var tags []string
		if len(tag) != 0 {
			tags = append(tags, tag)
		}
		err := c.comRegisterService(nTarget, proto, tags, address, port, dereg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GetServiceByDc(target, proto, tag, dc string) ([]string, uint64, error) {

	var address []string
	// target = strings.Replace(target+"."+proto, ".", "-", -1)
	if proto == "http" {
		target += "-http"
	} else {
		target += "-" + proto
	}

	qs := new(api.QueryOptions)

	if len(dc) != 0 {
		qs.Datacenter = dc
	}

	service, meta, err := c.client.Health().Service(target, tag, true, nil)
	for _, s := range service {
		address = append(address, s.Service.Address+":"+strconv.Itoa(s.Service.Port))
	}
	if err != nil {
		return nil, 0, err
	}
	return address, meta.LastIndex, nil
}

// GetService get the given service
func (c *Client) GetService(target, proto, tag string) ([]string, uint64, error) {
	return c.GetServiceByDc(target, proto, tag, "")
}

// WatchService watch the given service
func (c *Client) WatchService(target, proto, tag string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// target = strings.Replace(target+"."+proto, ".", "-", -1)
	if proto == "http" {
		target += "-http"
	} else {
		target += "-" + proto
	}

	// log.Debug("do watch service")

	respChan := make(chan watchResponse)
	go func() {

		defer func() {
			if err := recover(); err != nil {
				respChan <- watchResponse{waitIndex, errors.New(fmt.Sprintf("watchService panic,target:%v,proto:%v,tag:%v", target, proto, tag))}
			}
		}()

		opts := api.QueryOptions{
			WaitIndex: waitIndex,
			WaitTime:  20 * time.Second,
		}
		// log.Debug("start watch service ,target:",target,",waitIndex:",waitIndex)
		_, meta, err := c.client.Health().Service(target, tag, true, &opts)
		// log.Debug("watch end target:",target,",lastindex:",meta.LastIndex)
		if err != nil {
			respChan <- watchResponse{waitIndex, err}
			return
		}
		respChan <- watchResponse{meta.LastIndex, err}
	}()
	select {
	case <-stopChan:
		return waitIndex, nil
	case r := <-respChan:
		return r.waitIndex, r.err
	}

}
