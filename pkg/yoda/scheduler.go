package yoda

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Name = "yoda"
)

var (
	_ framework.PreFilterPlugin = &Yoda{}
	_ framework.FilterPlugin    = &Yoda{}
	_ framework.PreBindPlugin   = &Yoda{}

	scheme = runtime.NewScheme()
)

type Yoda struct {
	handle framework.Handle
}

func (y *Yoda) Name() string {
	return Name
}

func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	mgrConfig := ctrl.GetConfigOrDie()
	mgrConfig.QPS = 1000
	mgrConfig.Burst = 1000

	mgr, err := ctrl.NewManager(mgrConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "",
		LeaderElection:     false,
		Port:               9443,
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	go func() {
		if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			klog.Error(err)
			panic(err)
		}
	}()

	return &Yoda{
		handle: h,
	}, nil
}

func (y *Yoda) PreBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (y *Yoda) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	panic("implement me")
}

func (y *Yoda) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	panic("implement me")
}

func (y *Yoda) PreFilterExtensions() framework.PreFilterExtensions {
	panic("implement me")
}
