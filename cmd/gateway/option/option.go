// Copyright (C) 2014-2018 Goodrain Co., Ltd.
// RAINBOND, Application Management Platform

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package option

import (
	"fmt"
	"os"
	"time"

	"github.com/goodrain/rainbond/util"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
)

// GWServer contains Config and LogLevel
type GWServer struct {
	Config
	LogLevel string
	Debug    bool
}

// NewGWServer creates a new option.GWServer
func NewGWServer() *GWServer {
	return &GWServer{}
}

//Config contains all configuration
type Config struct {
	K8SConfPath  string
	EtcdEndpoint []string
	EtcdTimeout  int
	EtcdCaFile   string
	EtcdCertFile string
	EtcdKeyFile  string
	ListenPorts  ListenPorts
	//This number should be, at maximum, the number of CPU cores on your system.
	WorkerProcesses    int
	WorkerRlimitNofile int
	ErrorLog           string
	ErrorLogLevel      string
	WorkerConnections  int
	//essential for linux, optmized to serve many clients with each thread
	EnableEpool       bool
	EnableMultiAccept bool
	KeepaliveTimeout  int
	KeepaliveRequests int
	NginxUser         string
	ResyncPeriod      time.Duration
	// health check
	HealthPath         string
	HealthCheckTimeout time.Duration

	EnableMetrics bool

	NodeName        string
	HostIP          string
	IgnoreInterface []string
	ShareMemory     uint64
	SyncRateLimit   float32
}

// ListenPorts describe the ports required to run the gateway controller
type ListenPorts struct {
	HTTP   int
	HTTPS  int
	Status int
	Stream int
	Health int
}

// AddFlags adds flags
func (g *GWServer) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&g.LogLevel, "log-level", "debug", "the gateway log level")
	fs.StringVar(&g.K8SConfPath, "kube-conf", "", "absolute path to the kubeconfig file")
	fs.IntVar(&g.ListenPorts.Status, "status-port", 18080, `Port to use for the lua HTTP endpoint configuration.`)
	fs.IntVar(&g.ListenPorts.Stream, "stream-port", 18081, `Port to use for the lua TCP/UDP endpoint configuration.`)
	fs.IntVar(&g.ListenPorts.Health, "healthz-port", 10254, `Port to use for the healthz endpoint.`)
	fs.IntVar(&g.ListenPorts.HTTP, "service-http-port", 80, `Port to use for the http service rule`)
	fs.IntVar(&g.ListenPorts.HTTPS, "service-https-port", 443, `Port to use for the https service rule`)
	fs.IntVar(&g.WorkerProcesses, "worker-processes", 0, "Default get current compute cpu core number.This number should be, at maximum, the number of CPU cores on your system.")
	fs.IntVar(&g.WorkerConnections, "worker-connections", 4000, "Determines how many clients will be served by each worker process.")
	fs.IntVar(&g.WorkerRlimitNofile, "worker-rlimit-nofile", 200000, "Number of file descriptors used for Nginx. This is set in the OS with 'ulimit -n 200000'")
	fs.BoolVar(&g.EnableEpool, "enable-epool", true, "essential for linux, optmized to serve many clients with each thread")
	fs.BoolVar(&g.EnableMultiAccept, "enable-multi-accept", true, "Accept as many connections as possible, after nginx gets notification about a new connection.")
	fs.StringVar(&g.ErrorLog, "error-log", "/dev/stderr", "nginx log file, stderr or syslog")
	fs.StringVar(&g.ErrorLogLevel, "errlog-level", "crit", "log level")
	fs.StringVar(&g.NginxUser, "nginx-user", "root", "nginx user name")
	fs.IntVar(&g.KeepaliveRequests, "keepalive-requests", 10000, "Number of requests a client can make over the keep-alive connection. ")
	fs.IntVar(&g.KeepaliveTimeout, "keepalive-timeout", 30, "Timeout for keep-alive connections. Server will close connections after this time.")
	fs.DurationVar(&g.ResyncPeriod, "resync-period", 10*time.Second, "the default resync period for any handlers added via AddEventHandler and how frequently the listener wants a full resync from the shared informer")
	// etcd
	fs.StringSliceVar(&g.EtcdEndpoint, "etcd-endpoints", []string{"http://127.0.0.1:2379"}, "etcd cluster endpoints.")
	fs.IntVar(&g.EtcdTimeout, "etcd-timeout", 10, "etcd http timeout seconds")
	fs.StringVar(&g.EtcdCaFile, "etcd-ca", "", "etcd tls ca file ")
	fs.StringVar(&g.EtcdCertFile, "etcd-cert", "", "etcd tls cert file")
	fs.StringVar(&g.EtcdKeyFile, "etcd-key", "", "etcd http tls cert key file")
	// health check
	fs.StringVar(&g.HealthPath, "health-path", "/healthz", "absolute path to the kubeconfig file")
	fs.DurationVar(&g.HealthCheckTimeout, "health-check-timeout", 10, `Time limit, in seconds, for a probe to health-check-path to succeed.`)
	fs.BoolVar(&g.EnableMetrics, "enable-metrics", true, "Enables the collection of rbd-gateway metrics")
	fs.StringVar(&g.NodeName, "node-name", "", "this gateway node host name")
	fs.StringVar(&g.HostIP, "node-ip", "", "this gateway node ip")
	fs.BoolVar(&g.Debug, "debug", false, "enable pprof debug")
	fs.Uint64Var(&g.ShareMemory, "max-config-share-memory", 128, "Nginx maximum Shared memory size, which should be increased for larger clusters.")
	fs.Float32Var(&g.SyncRateLimit, "sync-rate-limit", 0.3, "Define the sync frequency upper limit")
	fs.StringArrayVar(&g.IgnoreInterface, "ignore-interface", []string{"docker0", "tunl0", "cni0", "kube-ipvs0", "flannel"}, "The network interface name that ignore by gateway")
}

// SetLog sets log
func (g *GWServer) SetLog() {
	level, err := logrus.ParseLevel(g.LogLevel)
	if err != nil {
		fmt.Println("set log level error." + err.Error())
		return
	}
	logrus.SetLevel(level)
}

//CheckConfig check config
func (g *GWServer) CheckConfig() error {
	if g.NodeName == "" {
		g.NodeName, _ = os.Hostname()
	}
	if g.HostIP == "" {
		ip, err := util.LocalIP()
		if err != nil {
			logrus.Errorf("get ip failed,details %s", err.Error())
			return fmt.Errorf("get host ip failure %s", err.Error())
		}
		g.HostIP = ip.String()
	}
	return nil
}
