package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/isayme/go-logger"
	"github.com/isayme/go-xlan/xlan/conf"
	"github.com/isayme/go-xlan/xlan/protocol"
	"github.com/isayme/go-xlan/xlan/util"
)

func main() {
	config := conf.Get()

	server := NewServer(config)
	err := server.ListenAndServer()
	if err != nil {
		logger.Panic(err)
	}
}

type Service struct {
	server *Server
	cfg    *conf.ServiceConfig
	l      net.Listener
}

func NewService(server *Server, cfg *conf.ServiceConfig) *Service {
	return &Service{
		server: server,
		cfg:    cfg,
	}
}

func (s *Service) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.cfg.RemotePort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Errorw("listen fail", "service", s.cfg.Name, "err", err)
		return err
	}

	s.l = l

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Warnf("[%s] accept fail, err: %s", s.cfg.Name, err.Error())
			continue
		}

		go s.handleConnection(conn)
	}
}

/*
 * 接收到 user 请求, 需要告知 xlan-s
 */
func (s *Service) handleConnection(conn net.Conn) {
	logger.Debugw("new connection", "address", conn.RemoteAddr().String())

	c, err := s.getConnectionForService()
	if err != nil {
		logger.Warnw("getConnectionForService fail", "err", err)
		return
	}
	logger.Debugw("getConnectionForService ok", "conn", conn.RemoteAddr().String())

	util.Proxy(conn, c)
}

func (s *Service) getConnectionForService() (net.Conn, error) {
	return s.server.getConnectionForService(s.cfg.Name)
}

type eventCallback func(net.Conn)

type Server struct {
	addr string

	controllConnectionMap sync.Map
	serviceMap            sync.Map

	l    net.Listener
	wire *protocol.Wire

	eventCB sync.Map
}

func NewServer(config *conf.Config) *Server {
	return &Server{
		addr: fmt.Sprintf("%s:%d", config.Server.Addr, config.Server.Port),
		wire: protocol.NewWire(),
	}
}

func (s *Server) startService(cfg conf.ServiceConfig) error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.l = l
	return nil
}

func (s *Server) ListenAndServer() error {
	addr := s.addr

	l, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Errorw("server listen fail", "err", err)
		return err
	}

	logger.Infow("listen ...", "addr", addr)
	s.l = l

	for {
		conn, err := s.l.Accept()
		if err != nil {
			logger.Warnw("accept fail", "err", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	logger.Debugw("new connection", "address", conn.RemoteAddr().String())

	{
		command, err := s.wire.ReadCommand(conn)
		if err != nil {
			logger.Warnw("read command fail", "err", err)
			return
		}
		logger.Debugw("new connection", "cmd", command.Command.String(), "id", command.ID)

		if command.Command == protocol.RegisterAsProxy {
			args := &protocol.ArgsForRegisterAsProxy{}
			err := json.Unmarshal(command.Args, args)
			if err != nil {
				logger.Warnw("parse ArgsForRegisterAsProxy fail", "err", err)
				return
			}

			if value, ok := s.eventCB.LoadAndDelete(args.ID); ok {
				if cb, ok := value.(eventCallback); ok {
					cb(conn)
				}
			}
			return
		} else if command.Command != protocol.RegisterAsControl {
			logger.Warnw("not support command", "cmd", command.Command)
			conn.Close()
			return
		}
	}

	defer conn.Close()

	for {
		command, err := s.wire.ReadCommand(conn)
		if err != nil {
			logger.Warnw("read command fail", "err", err)
			break
		}

		logger.Debugw("read command ok", "cmd", command.Command.String(), "id", command.ID)
		switch command.Command {
		case protocol.RegisterService:
			go s.handleRegisterService(conn, command)
		default:
			logger.Warnw("unkown command", "cmd", command.Command.String(), "id", command.ID)
		}
	}
}

func (s *Server) getConnectionForService(name string) (net.Conn, error) {
	var controlConn net.Conn

	if value, ok := s.controllConnectionMap.Load(name); ok {
		if v, ok := value.(net.Conn); ok {
			controlConn = v
		}
	}

	if controlConn == nil {
		return nil, fmt.Errorf("service(%s) not register", name)
	}

	cmdID := util.UUID()
	ch := make(chan net.Conn)
	var listener eventCallback = func(conn net.Conn) {
		ch <- conn
	}
	s.eventCB.Store(cmdID, listener)

	command := &protocol.Command{
		ID:      cmdID,
		Command: protocol.NewConnectionFromUser,
	}

	args := protocol.ArgsForNewConnectionFromUser{
		Name: name,
	}
	err := s.wire.WriteCommand(controlConn, command, args)
	if err != nil {
		return nil, err
	}

	select {
	case conn := <-ch:
		return conn, nil
	case <-time.After(time.Second * 30):
		return nil, fmt.Errorf("getconn timeout, cmd(%s)", command.ID)
	}
}

/*
 * xlan-c 向 xlan-s 注册新服务
 */
func (s *Server) handleRegisterService(conn net.Conn, command *protocol.Command) {
	serviceConfig := &conf.ServiceConfig{}
	err := json.Unmarshal(command.Args, serviceConfig)
	if err != nil {
		logger.Warnw("parse service config fail", "err", err)
		return
	}

	s.controllConnectionMap.Store(serviceConfig.Name, conn)

	err = s.createServiceAndServe(serviceConfig)
	if err != nil {
		logger.Warnw("createService fail", "err", err)
		return
	}
}

func (s *Server) createServiceAndServe(serviceConfig *conf.ServiceConfig) error {
	if _, ok := s.serviceMap.Load(serviceConfig.Name); ok {
		// exist
		return nil
	}

	service := NewService(s, serviceConfig)
	s.serviceMap.Store(serviceConfig.Name, service)
	err := service.ListenAndServe()
	if err != nil {
		logger.Warnw("service.ListenAndServe fail", "err", err)
	}

	return nil
}
