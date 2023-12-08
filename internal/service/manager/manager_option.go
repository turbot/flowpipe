package manager

// ManagerOption defines a type of function to configures the Manager.
type ManagerOption func(*Manager)

func WithESService() ManagerOption {
	return func(m *Manager) {
		m.startup |= startES
	}
}

func WithDocker() ManagerOption {
	return func(m *Manager) {
		m.startup |= startDocker
	}
}

func WithServerConfig(addr string, port int) ManagerOption {
	return func(m *Manager) {
		if addr == "local" {
			addr = "localhost"
		}
		m.HTTPAddress = addr
		m.HTTPPort = port
		m.startup |= startDocker | startES | startAPI | startScheduler
	}
}
