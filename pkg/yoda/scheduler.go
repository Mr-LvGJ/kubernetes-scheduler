package yoda

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
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
type Args struct {
	FavoriteColor  string `json:"favorite_color,omitempty"`
	FavoriteNumber int    `json:"favorite_number,omitempty"`
	ThanksTo       string `json:"thanks_to,omitempty"`
}
type Yoda struct {
	handle framework.Handle
}

func (y *Yoda) Name() string {
	return Name
}

func New(rargs runtime.Object, h framework.Handle) (framework.Plugin, error) {
	mgrConfig := ctrl.GetConfigOrDie()
	mgrConfig.QPS = 1000
	mgrConfig.Burst = 1000
	klog.V(3).Info("=======================================================")

	_, err := ctrl.NewManager(mgrConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "",
		LeaderElection:     false,
		Port:               9443,
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	//go func() {
	//	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
	//		klog.Error(err)
	//		panic(err)
	//	}
	//}()
	args := &Args{}
	if err := frameworkruntime.DecodeInto(rargs, args); err != nil {
		return nil,err
	}
	klog.V(3).Infof("get plugin config args: %+v", args)

	return &Yoda{
		handle: h,
	}, nil
}

func (y *Yoda) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.V(3).Infof("prefilter pod: %v", p.Name)
	return framework.NewStatus(framework.Success, "")
}

func (y *Yoda) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, nodeInfo.Node().GetName())
	return framework.NewStatus(framework.Success, "")
}

func (y *Yoda) PreBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	if nodeInfo, err := y.handle.SnapshotSharedLister().NodeInfos().Get(nodeName) ; err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("prebind get node info error: %+v", nodeName))
	} else {
		klog.V(3).Infof("prebind node info : %+v", nodeInfo.Node())
		return framework.NewStatus(framework.Success, "")
	}
}

func (y *Yoda) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}
