package manager

// ManagerOption defines a type of function to configures the Manager.
type ManagerOption func(*Manager)

func WithRaftNodeID(nodeID string) ManagerOption {
	return func(m *Manager) {
		m.RaftNodeID = nodeID
	}
}

func WithRaftBootstrap(bootstrap bool) ManagerOption {
	return func(m *Manager) {
		m.RaftBootstrap = bootstrap
	}
}

func WithRaftAddress(addr string) ManagerOption {
	return func(m *Manager) {
		m.RaftAddress = addr
	}
}

func WithESService() ManagerOption {
	return func(m *Manager) {
		m.startES = true
	}
}

func WithServerConfig(addr string, port int) ManagerOption {
	return func(m *Manager) {
		m.HTTPAddress = addr
		m.HTTPPort = port
		m.serverMode = true
		m.startES = true
	}
}
