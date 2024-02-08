package app

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServerOption struct {
	ListenAddress string
	Port          int
	ProbeAddr     string
	MetricsAddr   string
	CertDir       string
	CertName      string
	KeyName       string
	ClientCAName  string

	Scheme *runtime.Scheme
}

func NewServerOption() *ServerOption {
	return &ServerOption{}
}

func (s *ServerOption) WithScheme(scheme *runtime.Scheme) *ServerOption {
	s.Scheme = scheme
	return s
}

func (s *ServerOption) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&s.ListenAddress, "listen-address", "", "The address the webhook listen.")
	fs.IntVar(&s.Port, "port", 8443, "The port the webhook listen.")

	fs.StringVar(&s.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.StringVar(&s.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")

	fs.StringVar(&s.CertDir, "cert-dir", "/admission.local.config/certificates/", "tls cert dir")
	fs.StringVar(&s.CertName, "cert-name", "tls.crt", "tls ca file name")
	fs.StringVar(&s.KeyName, "key-name", "tls.key", "tls key file name")
	fs.StringVar(&s.ClientCAName, "client-ca-name", "", "tls client key, if not set,skip tls verify")

}
