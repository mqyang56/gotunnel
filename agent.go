package gotunnel

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"sigs.k8s.io/apiserver-network-proxy/pkg/agent"
)

func NewTunnel(cloudAddr string, Id string, stopCh <-chan struct{}) {
	cc := agent.ClientSetConfig{
		Address:          cloudAddr,
		AgentID:          Id,
		SyncInterval:     1 * time.Second,
		ProbeInterval:    1 * time.Second,
		AgentIdentifiers: fmt.Sprintf("host=%s", Id),
		DialOptions: []grpc.DialOption{grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                1 * time.Hour,
			PermitWithoutStream: false,
		})},
	}
	client := cc.NewAgentClientSet(stopCh)
	client.Serve()
	<-stopCh
}
