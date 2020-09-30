package consul

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/registry"
)

type consulWatcher struct {
	r  *consulRegistry
	wo registry.WatchOptions
	wp *watch.Plan

	next chan *registry.Result
	exit chan bool

	nodeMap map[string]*registry.Service // 一个服务下的，所有实例
}

func newConsulWatcher(cr *consulRegistry, opts ...registry.WatchOption) (registry.Watcher, error) {

	var wo registry.WatchOptions
	for _, o := range opts {
		o(&wo)
	}

	cw := &consulWatcher{
		r:  cr,
		wo: wo,

		next: make(chan *registry.Result),
		exit: make(chan bool),

		nodeMap: make(map[string]*registry.Service),
	}

	if len(cw.wo.Service) > 0 {
		// watch service
		if err := cw.watchService(cw.wo.Service); err != nil {
			return nil, err
		}
	} else {
		// watch all, 暂不实现
	}

	return cw, nil
}

func (cw *consulWatcher) watchService(serviceName string) error {

	wp, err := watch.Parse(map[string]interface{}{
		"type":    "service",
		"service": serviceName,
	})
	if err != nil {
		log.Error("err", "err", err)
		return err
	}

	wp.Handler = cw.serviceHandler
	cw.wp = wp

	go func() {
		wp.RunWithClientAndHclog(cw.r.Client(), nil)
		log.Info("RunWithClientAndHclog stop")
	}()

	log.Info("watch service", "serviceName", serviceName)
	return nil
}

// 针对某服务 变化的响应处理函数
func (cw *consulWatcher) serviceHandler(idx uint64, data interface{}) {

	entries, ok := data.([]*api.ServiceEntry)
	if !ok {
		return
	}

	newNodeMap := make(map[string]*registry.Service)
	curNodeMap := make(map[string]*registry.Service)
	for _, e := range entries {
		// 对于同一个服务来说， 地址+端口 可以作为唯一的一个实例
		key := e.Service.Address + "-" + strconv.Itoa(e.Service.Port)

		svc, ok := cw.nodeMap[key]
		if !ok {
			svc = &registry.Service{
				Endpoints: nil,
				Name:      e.Service.Service,
				Version:   "",
				Nodes:     make([]*registry.Node, 0, 1),
			}

			svc.Nodes = append(svc.Nodes, &registry.Node{
				Id:       e.Service.ID,
				Address:  fmt.Sprintf("%s:%d", e.Service.Address, e.Service.Port), // ip地址和端口
				Metadata: make(map[string]string),
			})

			// 新的实例
			cw.nodeMap[key] = svc
			newNodeMap[key] = svc
		}

		curNodeMap[key] = svc
	}

	deletedMap := make(map[string]*registry.Service)
	for key, v := range cw.nodeMap {
		if _, ok := curNodeMap[key]; !ok {
			deletedMap[key] = v
		}
	}

	for key, delService := range deletedMap {
		delete(cw.nodeMap, key)
		cw.next <- &registry.Result{Action: "delete", Service: delService}
	}

	for _, newService := range newNodeMap {
		cw.next <- &registry.Result{Action: "create", Service: newService}
	}
}

// 实现 registry.Watcher 中的 Next
func (cw *consulWatcher) Next() (*registry.Result, error) {

	select {
	case <-cw.exit:
		return nil, registry.ErrWatcherStopped
	case r, ok := <-cw.next:
		if !ok {
			return nil, registry.ErrWatcherStopped
		}
		return r, nil
	}
	// NOTE: This is a dead code path: e.g. it will never be reached
	// as we return in all previous code paths never leading to this return
	return nil, registry.ErrWatcherStopped
}

// 实现 registry.Watcher 中的 Stop
func (cw *consulWatcher) Stop() {

	select {
	case <-cw.exit:
		// 既然已经退出了，那就不需要重复执行了
		return
	default:
		close(cw.exit)
		if cw.wp == nil {
			return
		}
		cw.wp.Stop()

		// drain results
		for {
			select {
			case <-cw.next:
			default:
				return
			}
		}
	}
}
