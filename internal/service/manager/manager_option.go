package manager

// ManagerOption defines a type of function to configures the Manager.
type ManagerOption func(*Manager)

func WithESService() ManagerOption {
	return func(m *Manager) {
		m.startup |= startES
	}
}

func WithServerConfig(addr string, port int) ManagerOption {
	return func(m *Manager) {

		// map addr values as follows:
		//		"local" -> "localhost"
		//      "network" -> "" , i.e. listen on all addresses
		switch addr {
		case "local", "localhost":
			// listen on local host
			m.HTTPAddress = "localhost"
		case "network":
			// listen on all addresses
			addr = ""
		default:
			m.HTTPAddress = addr

		}
		m.HTTPPort = port
		m.startup |= startES | startAPI | startScheduler
	}
}
