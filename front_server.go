package gotunnel

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-network-proxy/pkg/server"
	"sigs.k8s.io/apiserver-network-proxy/proto/agent"
)

const (
	maxMsgSize = 200 * 1024 * 1024
)

type FrontServer struct {
	frontAddr     string
	agentServer   *grpc.Server
	agentListener net.Listener
	frontListener net.Listener
}

func NewFrontServer(frontAddr string) (*FrontServer, error) {
	frontServer := &FrontServer{}
	proxyServer := server.NewProxyServer(uuid.New().String(), []server.ProxyStrategy{server.ProxyStrategyDestHost}, 1, &server.AgentTokenAuthenticationOptions{})
	ka := keepalive.ServerParameters{
		Time: 1 * time.Hour,
	}

	// a:=proxyServer.BackendManagers[0].(*server.DestHostBackendManager)
	// a.ListAgentIds()

	frontServer.agentServer = grpc.NewServer(
		grpc.KeepaliveParams(ka),
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize))
	agent.RegisterAgentServiceServer(frontServer.agentServer, proxyServer)
	agentListener, err := net.Listen("tcp", frontAddr)
	if err != nil {
		return nil, err
	}
	frontServer.agentListener = agentListener
	go func() {
		klog.Infof("Start agent %s for edge node", frontAddr)
		err = frontServer.agentServer.Serve(frontServer.agentListener)
		if err != nil {
			klog.Fatal(err)
		}
	}()

	// http tunnel
	httpServer := &http.Server{
		Handler: &server.Tunnel{
			Server: proxyServer,
		},
	}

	frontListener, err := net.Listen("tcp", "")
	if err != nil {
		return nil, err
	}
	frontServer.frontListener = frontListener
	go func() {
		err := httpServer.Serve(frontServer.frontListener)
		if err != nil {
			klog.Fatal(err)
		}
	}()

	frontServer.frontAddr = frontServer.frontListener.Addr().String()
	klog.Infof("---------------> front listen: %s", frontServer.frontAddr)
	return frontServer, nil
}

func (s *FrontServer) Close() {
	s.agentServer.Stop()
	s.agentListener.Close()
	s.frontListener.Close()
}

func (s *FrontServer) DoRequest(nodeId, nodePort string, req *http.Request) (resp *http.Response, err error) {
	conn, err := net.Dial("tcp", s.frontAddr)
	if err != nil {
		klog.Errorf("Failed to Dial: %v", err)
		return nil, err
	}

	dialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return conn, nil
	}

	klog.Infof("conn: local %s-> remote %s ", conn.LocalAddr(), conn.RemoteAddr())
	client := http.Client{Transport: &http.Transport{
		DialContext: dialer,
	}}

	// Send HTTP-Connect request
	_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", fmt.Sprintf("%s:%s", nodeId, nodePort), nodeId)
	if err != nil {
		klog.Errorf("Failed to Fprintf: %v", err)
		return nil, err
	}

	// Parse the HTTP response for Connect
	br := bufio.NewReader(conn)
	res, err := http.ReadResponse(br, nil)
	if err != nil {
		klog.Errorf("Failed to ReadResponse: %v", err)
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(res.Status)
	}
	if br.Buffered() > 0 {
		return nil, fmt.Errorf("buffered size >0")
	}

	resp, err = client.Do(req)
	if err != nil {
		klog.Errorf("Failed to Do: %v", err)
		return nil, err
	}
	return
}
