module github.com/mqyang56/gotunnel

go 1.16

require (
	github.com/google/uuid v1.3.0
	google.golang.org/grpc v1.48.0
	k8s.io/klog/v2 v2.70.1
	sigs.k8s.io/apiserver-network-proxy v0.0.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	golang.org/x/net v0.0.0-20220805013720-a33c5aa5df48 // indirect
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect
	google.golang.org/genproto v0.0.0-20220805133916-01dd62135a58 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace (
	sigs.k8s.io/apiserver-network-proxy => github.com/mqyang56/apiserver-network-proxy v0.0.0-20220810032623-50ce1a99a496
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client => github.com/mqyang56/apiserver-network-proxy/konnectivity-client v0.0.0-20220810032623-50ce1a99a496
)
