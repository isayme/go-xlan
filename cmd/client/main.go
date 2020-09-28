package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/isayme/go-logger"
	"github.com/isayme/go-xlan/xlan/conf"
	"github.com/isayme/go-xlan/xlan/protocol"
	"github.com/isayme/go-xlan/xlan/util"
	"golang.org/x/sync/errgroup"
)

func main() {
	config := conf.Get()
	client := NewClient(config)

	for {
		err := client.StartAndServe()
		logger.Info("client.StartAndServe error", "err", err)
		time.Sleep(time.Second * 10)
		logger.Infof("reconnecting %s", config.Client.Server.Addr)
	}
}

type Client struct {
	services   []conf.ServiceConfig
	serviceMap map[string]string
	cfg        *conf.ClientConfig

	wire *protocol.Wire
}

func NewClient(config *conf.Config) *Client {
	return &Client{
		services:   config.Services,
		cfg:        &config.Client,
		serviceMap: make(map[string]string),
		wire:       protocol.NewWire(),
	}
}

func (c *Client) StartAndServe() error {
	conn, err := c.connectServer()
	if err != nil {
		return err
	}

	{
		command := &protocol.Command{
			Command: protocol.RegisterAsControl,
		}
		err = c.wire.WriteCommand(conn, command)
		if err != nil {
			logger.Errorw("register as control fail", "err", err)
			return err
		}
	}

	for _, serviceConfig := range c.services {
		c.serviceMap[serviceConfig.Name] = fmt.Sprintf("%s:%d", serviceConfig.LocalIP, serviceConfig.LocalPort)
		command := &protocol.Command{
			Command: protocol.RegisterService,
		}
		args := &protocol.ArgsForRegisterService{
			ServiceConfig: serviceConfig,
		}
		logger.Debugw("register service", "name", serviceConfig.Name)
		err = c.wire.WriteCommand(conn, command, args)
		if err != nil {
			logger.Warnw("register service fail", "name", serviceConfig.Name, "err", err)
			return err
		}
	}

	for {
		command, err := c.wire.ReadCommand(conn)
		if err != nil {
			logger.Errorw("read command fail", "err", err)
			break
		}
		logger.Debugw("read command", "cmd", command.Command.String(), "id", command.ID)

		switch command.Command {
		case protocol.NewConnectionFromUser:
			go c.handleNewConnectionFromUser(command)
		default:
			logger.Warnw("unkown command", "cmd", command.Command.String(), "id", command.ID)
		}
	}

	return nil
}

func (c *Client) connectServer() (net.Conn, error) {
	severAddr := fmt.Sprintf("%s:%d", c.cfg.Server.Addr, c.cfg.Server.Port)
	conn, err := net.Dial("tcp", severAddr)
	if err != nil {
		logger.Warnw("dial server fail", "err", err)
		return nil, err
	}

	return conn, nil
}

func (c *Client) getServiceAddrByName(name string) string {
	addr, _ := c.serviceMap[name]
	return addr
}

// 同时向 xlan-s 和 server 连接
func (c *Client) handleNewConnectionFromUser(command *protocol.Command) {
	argsForConnectionFromUser := &protocol.ArgsForNewConnectionFromUser{}
	err := json.Unmarshal(command.Args, argsForConnectionFromUser)
	if err != nil {
		logger.Warnw("argsForConnectionFromUser: parse service name fail", "err", err)
		return
	}

	g, _ := errgroup.WithContext(context.Background())

	var serverConn net.Conn
	var serviceConn net.Conn

	defer func() {
		if serverConn != nil {
			serverConn.Close()
		}
		if serviceConn != nil {
			serviceConn.Close()
		}
	}()

	g.Go(func() error {
		var err error
		serverConn, err = c.connectServer()
		if err != nil {
			return err
		}

		registerAsProxyCommand := &protocol.Command{
			Command: protocol.RegisterAsProxy,
		}
		args := protocol.ArgsForRegisterAsProxy{
			ID: command.ID,
		}
		err = c.wire.WriteCommand(serverConn, registerAsProxyCommand, args)
		if err != nil {
			logger.Warn("register proxy conn fail", "err", err)
			return err
		}

		return nil
	})
	g.Go(func() error {
		var err error
		serviceAddr := c.getServiceAddrByName(argsForConnectionFromUser.Name)
		serviceConn, err = net.Dial("tcp", serviceAddr)
		if err != nil {
			logger.Warnw("dial service fail", "err", err)
			return err
		}

		return nil
	})

	err = g.Wait()
	if err != nil {
		logger.Warnw("create proxy fail", "err", err)
		return
	}

	util.Proxy(serverConn, serviceConn)
}
