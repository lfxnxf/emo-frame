package registry

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/tpc/inf/go-upstream/config"
)

type ServiceManager struct {
	logger      *logging.Logger
	mutex       *sync.Mutex
	regs        map[string]*config.Register
	closed      bool
	serviceTags []string
}

func NewServiceManager(logger *logging.Logger) *ServiceManager {
	return &ServiceManager{
		mutex:  new(sync.Mutex),
		closed: false,
		logger: logger,
		regs:   make(map[string]*config.Register),
	}
}

func convertMapToStringSlice(m map[string]string) []string {
	data := make([]string, 0, len(m))
	for k, v := range m {
		data = append(data, fmt.Sprintf("%s=%s", k, v))

	}
	sort.Sort(sort.StringSlice(data))
	return data
}

func getServiceInstanceWatchPath(prefix, addr string, port int) string {
	return prefix + "/" + addr + "/" + strconv.Itoa(port)
}

func (bm *ServiceManager) Register(reg *config.Register) error {
	// 存在则忽略
	key := reg.ServiceName + "-" + reg.ServiceAddr + "-" + strconv.Itoa(reg.ServicePort)
	bm.mutex.Lock()
	if _, ok := bm.regs[key]; ok {
		bm.mutex.Unlock()
		return nil
	}
	bm.mutex.Unlock()

	var parse = func(tag string) (map[string]string, error) {
		tagMap := make(map[string]string)
		err := json.Unmarshal([]byte(tag), &tagMap)
		if err != nil {
			bm.logger.Warnf("parse service tags error %s", err)
			return nil, err
		}
		return tagMap, nil
	}

	var override = func(origin, new map[string]string) map[string]string {
		data := make(map[string]string)
		for k, v := range origin {
			data[k] = v
		}
		for k, v := range new {
			data[k] = v
		}
		return data
	}

	var dynamicTags map[string]string
	watchPath := getServiceInstanceWatchPath(reg.TagsWatchPath, reg.ServiceAddr, reg.ServicePort)
	dynamicTagsStr, _, err := Default.ReadManual(watchPath)
	if len(dynamicTagsStr) != 0 {
		dynamicTags, err = parse(dynamicTagsStr)
		if err != nil {
			bm.logger.Warnf("parse service tags error %s", err)
			return err
		}
		dynamicTags = override(reg.ServiceTags, dynamicTags)
	} else {
		dynamicTags = override(reg.ServiceTags, nil)
	}
	reg.TagsOverrideCh <- convertMapToStringSlice(dynamicTags)
	bm.logger.Infof("consul: service register with remote tags override local %q, remote %q, path %q", strings.Join(convertMapToStringSlice(reg.ServiceTags), ","), strings.Join(convertMapToStringSlice(dynamicTags), ","), watchPath)
	err = Default.Register(reg)
	if err != nil {
		return err
	}
	// 再次check
	key = reg.ServiceName + "-" + reg.ServiceAddr + "-" + strconv.Itoa(reg.ServicePort)
	bm.mutex.Lock()
	if _, ok := bm.regs[key]; ok {
		bm.mutex.Unlock()
		return nil
	}
	bm.regs[key] = reg
	bm.mutex.Unlock()
	var reload = func() {
		tagsCh := Default.WatchManual(watchPath)
		for tag := range tagsCh {
			if tag == dynamicTagsStr {
				continue
			}
			dynamicTagsStr = tag
			dynamicTags, err = parse(tag)
			if err != nil {
				continue
			}
			dynamicTags = override(reg.ServiceTags, dynamicTags)
			tagSlice := convertMapToStringSlice(dynamicTags)
			bm.logger.Infof("consul watch service tags changed, tags override new %q", strings.Join(tagSlice, ","))

			bm.mutex.Lock()
			bm.serviceTags = make([]string, len(tagSlice))
			copy(bm.serviceTags, tagSlice)
			bm.mutex.Unlock()

			reg.TagsOverrideCh <- tagSlice
		}
	}
	go reload()
	return nil
}

func (bm *ServiceManager) Deregister() {
	bm.mutex.Lock()
	for k, reg := range bm.regs {
		reg := reg
		Default.Deregister(reg)
		delete(bm.regs, k)
	}
	bm.mutex.Unlock()
}

func (bm *ServiceManager) ServiceTags() []string {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	rst := make([]string, len(bm.serviceTags))
	copy(rst, bm.serviceTags)
	return rst
}

func (bm *ServiceManager) InitServiceTags(sdname string, port int) {
	namespace := strings.Split(sdname, ".")[0]
	path := fmt.Sprintf("/service_config/%s/%s/service_tags", namespace, sdname)
	watchPath := getServiceInstanceWatchPath(path, config.LocalIPString(), port)
	dynamicTagsStr, _, err := Default.ReadManual(watchPath)
	if err != nil {
		bm.logger.Warnf("consul: Error read service tags, path %s, error %s", path, err)
	}
	var parse = func(tag string) (map[string]string, error) {
		tagMap := make(map[string]string)
		if len(tag) == 0 {
			return tagMap, nil
		}
		err := json.Unmarshal([]byte(tag), &tagMap)
		if err != nil {
			bm.logger.Warnf("parse service tags error %s", err)
			return nil, err
		}
		return tagMap, nil
	}
	dynamicTags, _ := parse(dynamicTagsStr)
	s := convertMapToStringSlice(dynamicTags)
	bm.mutex.Lock()
	bm.serviceTags = make([]string, len(s))
	copy(bm.serviceTags, s)
	bm.mutex.Unlock()
}

func (bm *ServiceManager) SetLogger(l *logging.Logger) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	bm.logger = l
}
