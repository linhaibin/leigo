package lei

const (
	DEFAULT_MAX_READ_MSG_SIZE    int = 1024 * 1024
	DEFAULT_READ_MSG_QUEUE_SIZE  int = 10 * 1024
	DEFAULT_READ_TIME_OUT        int = 0
	DEFAULT_MAX_WRITE_MSG_SIZE   int = 1024 * 1024
	DEFAULT_WRITE_MSG_QUERE_SIZE int = 10 * 1024
	DEFAULT_WRITE_TIME_OUT       int = 0
)

type Config struct {
	Address string //ip:port
	//read config
	MaxReadMsgSize   int
	ReadMsgQueueSize int
	ReadTimeOut      int
	//write config
	MaxWriteMsgSize   int
	WriteMsgQueueSize int
	WriteTimeOut      int
}

func (this *Config) Check() bool {
	if this.MaxReadMsgSize == 0 {
		this.MaxReadMsgSize = DEFAULT_MAX_READ_MSG_SIZE
	}
	if this.ReadMsgQueueSize == 0 {
		this.ReadMsgQueueSize = DEFAULT_READ_MSG_QUEUE_SIZE
	}
	if this.ReadTimeOut == 0 {
		this.ReadTimeOut = DEFAULT_READ_TIME_OUT
	}
	if this.MaxWriteMsgSize == 0 {
		this.MaxWriteMsgSize = DEFAULT_MAX_WRITE_MSG_SIZE
	}
	if this.WriteMsgQueueSize == 0 {
		this.WriteMsgQueueSize = DEFAULT_WRITE_MSG_QUERE_SIZE
	}
	if this.WriteTimeOut == 0 {
		this.WriteTimeOut = DEFAULT_WRITE_TIME_OUT
	}
	return true
}
