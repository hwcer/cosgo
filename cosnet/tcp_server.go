package cosnet

import (
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/storage"
	"net"
	"time"
)

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}

func NewTcpServer(address string, handler Handler) *TcpServer {
	s := &TcpServer{
		NetServer: NewNetServer(address, handler, MsgTypeMsg, NetTypeTcp),
	}
	s.sockets.Array.NewDataset = s.addSocket
	return s
}

func (s *TcpServer) Close() error {
	if !s.SCC.Close() {
		return nil
	}
	s.listener.Close()
	return s.SCC.Wait(time.Second * 10)
}

func (s *TcpServer) Start() error {
	if err := s.NetServer.Start(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.listener = listener
	s.GO(s.listen)
	return nil
}

func (s *TcpServer) listen() {
	//defer s.SCC.Cancel()
	for !s.Stopped() {
		c, err := s.listener.Accept()
		if err != nil {
			//logger.Error("tcp server accept failed:%v", err)
			break
		} else {
			go s.socket(c)
		}
	}
}

func (s *TcpServer) socket(conn net.Conn) Socket {
	sock := &TcpSocket{
		conn:      conn,
		NetSocket: NewNetSocket(s),
	}
	s.sockets.Set(sock)
	s.Emit(EventsTypeConnect, sock)
	s.GO(sock.readMsg)
	s.GO(sock.writeMsg)
	logger.Debug("new socket Id:%d from Addr:%s", sock.Id(), sock.RemoteAddr())
	return sock
}

func (s *TcpServer) addSocket(id uint64, val interface{}) storage.ArrayDataset {
	sock := val.(*TcpSocket)
	sock.NetSocket.ArrayDatasetDefault = storage.NewArrayDataset(id, nil).(*storage.ArrayDatasetDefault)
	return sock
}
