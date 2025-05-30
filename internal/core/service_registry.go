package core

// ServiceRegistry 服务注册器 - 管理AI任务分发器
type ServiceRegistry struct {
	dispatchers map[string]AITaskDispatcher
}

// NewServiceRegistry 创建新的服务注册器
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		dispatchers: make(map[string]AITaskDispatcher),
	}
}

// RegisterDispatcher 注册AI任务分发器
func (sr *ServiceRegistry) RegisterDispatcher(dispatcher AITaskDispatcher) {
	sr.dispatchers[dispatcher.GetProviderName()] = dispatcher
}

// GetDispatcher 获取指定名称的AI任务分发器
func (sr *ServiceRegistry) GetDispatcher(name string) (AITaskDispatcher, bool) {
	dispatcher, exists := sr.dispatchers[name]
	return dispatcher, exists
}

// GetAllDispatchers 获取所有已注册的AI任务分发器
func (sr *ServiceRegistry) GetAllDispatchers() map[string]AITaskDispatcher {
	return sr.dispatchers
}

// ListProviderNames 获取所有已注册的服务提供商名称列表
func (sr *ServiceRegistry) ListProviderNames() []string {
	names := make([]string, 0, len(sr.dispatchers))
	for name := range sr.dispatchers {
		names = append(names, name)
	}
	return names
}

// HasProvider 检查是否已注册指定的服务提供商
func (sr *ServiceRegistry) HasProvider(name string) bool {
	_, exists := sr.dispatchers[name]
	return exists
}

// UnregisterDispatcher 注销AI任务分发器
func (sr *ServiceRegistry) UnregisterDispatcher(name string) bool {
	if _, exists := sr.dispatchers[name]; exists {
		delete(sr.dispatchers, name)
		return true
	}
	return false
}

// Count 获取已注册的分发器数量
func (sr *ServiceRegistry) Count() int {
	return len(sr.dispatchers)
}
