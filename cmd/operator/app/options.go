package app

import (
	"github.com/spf13/pflag"
)

type ServerOption struct {
	MetricsAddr          string
	ProbeAddr            string
	EnableLeaderElection bool
}

func NewServerOption() *ServerOption {
	return &ServerOption{}
}

func (s *ServerOption) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&s.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&s.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&s.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

}
