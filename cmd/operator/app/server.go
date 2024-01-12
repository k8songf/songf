package app

import (
	"fmt"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"songf.sh/songf/internal/controller"
)

func Run(opt ServerOption) error {

	controller.InitializeCache()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 opt.Scheme,
		Metrics:                metricsserver.Options{BindAddress: opt.MetricsAddr},
		HealthProbeBindAddress: opt.ProbeAddr,
		LeaderElection:         opt.EnableLeaderElection,
		LeaderElectionID:       "a8278f16.songf.sh",
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

	reconciler, err := controller.NewJobReconciler(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		return err
	}

	if err = reconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("%s:%s-%s-%s", err.Error(), "unable to create controller", "controller", "Job")
	}
	if err = (&controller.JobBatchReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("%s:%s-%s-%s", err.Error(), "unable to create controller", "controller", "JobBatch")
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "unable to set up health check")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "unable to set up ready check")
	}

	klog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("%s:%s", err.Error(), "problem running manager")
	}

	return nil

}
