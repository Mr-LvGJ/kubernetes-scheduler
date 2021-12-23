package yoda

import (
	"context"
	"errors"
	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	scv "github.com/NJUPT-ISL/SCV/api/v1"
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
	cache  cache.Cache
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

func (y *Yoda) Name() string {
	return Name
}

func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	mgrConfig := ctrl.GetConfigOrDie()
	mgrConfig.QPS = 1000
	mgrConfig.Burst = 1000

	if err := scv.AddToScheme(scheme); err != nil {
		klog.Error(err)
		return nil, err
	}

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

	scvCache := mgr.GetCache()

	if scvCache.WaitForCacheSync(context.TODO()) {
		return &Yoda{
			handle: h,
			cache:  scvCache,
		}, nil
	} else {
		return nil, errors.New("Cache Not Sync! ")
	}
}
