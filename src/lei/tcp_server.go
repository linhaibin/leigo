package lei

import (
	l4g "base/log4go"
	//"fmt"
	//"io"
	"net"
)

type Server interface {
	Close()
	NewSession() Sessioner
}

func TcpServe(s Server, conf *Config) {
	defer s.Close()

	l, err := net.Listen("tcp", conf.Address)
	if err != nil {
		l4g.Error("Listen on %s failed ! err:%v", conf.Address, err)
		return
	}
	l4g.Info("Listener is on.")
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			//Temporary error是什么意思？什么情况下会返回这个错误？
			//write和read也有可能会返回这个错误吗？
			if nerr, ok := err.(net.Error); !ok || nerr.Timeout() || nerr.Temporary() {
				continue
			}
			l4g.Error("Accept err ! err:%v", err)
			return
		}

		go newBroker(conf, s.NewSession()).serve(conn)
	}
}
