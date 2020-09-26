module cosgo

go 1.14

replace github.com/hwclegend/cosgo v0.0.0 => ../cosgo

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/hjson/hjson-go v3.0.1+incompatible
	github.com/hwclegend/cosgo v0.0.0 // indirect
	github.com/mitchellh/mapstructure v1.3.3
	github.com/montanaflynn/stats v0.6.3
	github.com/shirou/gopsutil v2.20.7+incompatible
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.4.0
	github.com/yuin/gopher-lua v0.0.0-20200603152657-dc2b0ca8b37e
	go.etcd.io/etcd v3.3.22+incompatible
	golang.org/x/sys v0.0.0-20200116001909-b77594299b42
)
