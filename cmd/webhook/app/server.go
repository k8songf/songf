package app

import (
	"fmt"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	hook "songf.sh/songf/internal/webhook"
)

func Run(opt ServerOption) error {

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 opt.Scheme,
		Metrics:                metricsserver.Options{BindAddress: opt.MetricsAddr},
		HealthProbeBindAddress: opt.ProbeAddr,
		LeaderElection:         false,

		WebhookServer: webhook.NewServer(webhook.Options{
			Host:         opt.ListenAddress,
			Port:         opt.Port,
			CertDir:      opt.CertDir,
			CertName:     opt.CertName,
			KeyName:      opt.KeyName,
			ClientCAName: opt.ClientCAName,
		}),
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "unable to start manager")
	}

	hookImpl, err := hook.NewJobWebHook()
	if err != nil {
		return err
	}

	if err = hookImpl.SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("%s:%s-%s-%s", err.Error(), "unable to create webhook", "webhook", "Job")
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "unable to set up health check")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "unable to set up ready check")
	}

	klog.Info("starting webhook")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "problem running manager")
	}

	return nil

}
