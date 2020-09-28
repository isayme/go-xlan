package protocol

import (
	"encoding/json"
	"net"

	pack "github.com/isayme/go-pack"
	"github.com/isayme/go-xlan/xlan/conf"
	"github.com/isayme/go-xlan/xlan/util"
)

type CommandType uint32

const (
	Noop                  CommandType = iota
	RegisterAsControl                 // xlan-c 向 xlan-s 建立连接, 标记自己身份为控制连接
	RegisterAsProxy                   // xlan-c 向 xlan-s 建立连接, 标记自己身份为代理连接
	RegisterService                   // xlan-c向xlan-s注册服务
	NewConnectionFromUser             // xlan-s 向 xlan-c 发送命令, 告知有新的 user 请求
)

func (ct CommandType) String() string {
	switch ct {
	case RegisterAsControl:
		return "RegisterAsControl"
	case RegisterAsProxy:
		return "RegisterAsProxy"
	case RegisterService:
		return "RegisterService"
	case NewConnectionFromUser:
		return "NewConnectionFromUser"
	default:
		return "unkown"
	}
}

type Command struct {
	ID      string      `json:"id,omitempty"`
	Command CommandType `json:"cmd"`
	Args    []byte      `json:"args,omitempty"`
}

type ArgsForRegisterAsProxy struct {
	ID string `json:"string"` // 等于请求的command.ID
}

// 注册service
type ArgsForRegisterService struct {
	conf.ServiceConfig
}

type ArgsForNewConnectionFromUser struct {
	Name string `json:"name"` // 服务名
}

type ArgsNewConnectionFromClient struct {
	ID string `json:"id"` // 响应请求ID
}

type Wire struct {
}

func NewWire() *Wire {
	return &Wire{}
}

func (w *Wire) ReadCommand(conn net.Conn) (*Command, error) {
	commandBuf, err := pack.Unpack(conn)

	command := &Command{}
	err = json.Unmarshal(commandBuf, command)
	if err != nil {
		return nil, err
	}

	return command, nil
}

func (w *Wire) WriteCommand(conn net.Conn, command *Command, args ...interface{}) error {
	if len(args) > 0 {
		argsBuf, err := json.Marshal(args[0])
		if err != nil {
			return err
		}

		command.Args = argsBuf
	}

	if command.ID == "" {
		command.ID = util.UUID()
	}

	buf, err := json.Marshal(command)
	if err != nil {
		return err
	}

	err = pack.Pack(conn, buf)
	if err != nil {
		return err
	}

	return nil
}
