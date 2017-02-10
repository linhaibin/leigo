package lei

func TcpClientServe(ses Sessioner, conf *Config) bool {
	return newBroker(conf, ses).Connect()
}
