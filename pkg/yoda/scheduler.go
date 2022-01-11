package yoda

import (
	"context"
	"fmt"
	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/advisor"
	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/cache"
	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/filter"
	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/score"
	"github.com/go-redis/redis/v8"
	"sync"

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
	_ framework.PreScorePlugin  = &Yoda{}
	_ framework.ScorePlugin     = &Yoda{}
	_ framework.ScoreExtensions = &Yoda{}
	_ framework.PreBindPlugin   = &Yoda{}

	scheme = runtime.NewScheme()
)

type Args struct {
	FavoriteColor  string `json:"favorite_color,omitempty"`
	FavoriteNumber int    `json:"favorite_number,omitempty"`
	ThanksTo       string `json:"thanks_to,omitempty"`
}
type Yoda struct {
	handle      framework.Handle
	resToWeight resourceToWeightMap
	redisClient *redis.Client
	mx          sync.RWMutex
}
type resourceToWeightMap map[v1.ResourceName]int64
type resourceToValueMap map[v1.ResourceName]int64

func (y *Yoda) Name() string {
	return Name
}

var res = &advisor.Result{}
var Resources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory, v1.ResourcePods, v1.ResourceStorage, v1.ResourceEphemeralStorage}

func New(rargs runtime.Object, h framework.Handle) (framework.Plugin, error) {
	mgrConfig := ctrl.GetConfigOrDie()
	mgrConfig.QPS = 1000
	mgrConfig.Burst = 1000

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

	resToWeightMap := make(resourceToWeightMap)
	redisClient := cache.New()
	for _, resource := range Resources {
		resToWeightMap[resource] = 1
	}

	args := &Args{}
	if err := frameworkruntime.DecodeInto(rargs, args); err != nil {
		return nil, err
	}
	klog.V(3).Infof("get plugin config args: %+v", args)
	return &Yoda{
		handle:      h,
		resToWeight: resToWeightMap,
		redisClient: redisClient,
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

func (y *Yoda) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	klog.V(3).Infof("collect info for scheduling pod: %v", pod.Name)
	y.redisClient.FlushDB(ctx)
	nodeInfo, err := res.Init()
	state.Write("nodeInfo", &nodeInfo)
	if err != nil {
		klog.Errorf("Get node info Error: %v", err)
		return framework.NewStatus(framework.Error, err.Error())
	}
	for info := range nodeInfo.Info {
		klog.V(3).Infof("info %v : %v", info, nodeInfo.Info[info])
	}
	return framework.NewStatus(framework.Success, "")
}

func (y *Yoda) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.V(3).Infof("Score pod: %v, node: %v", p.GetName(), nodeName)
	nodeInfo, err := y.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}
	nodeList, err := y.handle.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting nodelist err"))
	}
	currentNodeInfos, err := res.Init()
	currentNodeInfo := currentNodeInfos.Info[nodeName]
	klog.V(3).Infof("current node info : %v", currentNodeInfo)
	t := &advisor.Result{
		Info: currentNodeInfos.Info,
	}
	klog.V(3).Infof("ttttttttttttttttttt: %v", t.Info[nodeName])
	if err != nil {
		return 0, framework.NewStatus(framework.Error, "failed to get currentInfo")
	}
	requested := make(resourceToValueMap)
	allocatable := make(resourceToValueMap)

	for resource := range y.resToWeight {
		alloc, req := score.CalculateResourceAllocatableRequest(nodeInfo, p, resource, true)
		if alloc != 0 {
			// Only fill the extended resource entry when it's non-zero.
			allocatable[resource], requested[resource] = alloc, req
		}
		klog.V(3).Infof("resouce: %v, allocatable: %v, request: %v", resource, alloc, req)
	}
	y.mx.Lock()
	uNodeScore, err := score.CalculateScore(t, state, p, nodeInfo, nodeList, y.redisClient, allocatable)
	y.mx.Unlock()
	if err != nil {
		klog.Errorf("CalculateScore Error: %v", err)
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Score Node: %v Error: %v", nodeInfo.Node().Name, err))
	}
	nodeScore := filter.Uint64ToInt64(uNodeScore)
	return nodeScore, framework.NewStatus(framework.Success, "")
}

func (y *Yoda) NormalizeScore(ctx context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	klog.V(3).Infof("nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn")
	y.redisClient.FlushDB(ctx)
	var (
		highest int64 = 0
		lowest        = scores[0].Score
	)
	for _, nodeScore := range scores {
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
	}
	if highest == lowest {
		lowest--
	}

	for i, nodeScore := range scores {
		scores[i].Score = (nodeScore.Score - lowest) * framework.MaxNodeScore / (highest - lowest)
		klog.V(3).Infof("Node: %v, Score: %v in Plugin: Yoda When scheduling Pod: %v/%v", scores[i].Name, scores[i].Score, p.GetNamespace(), p.GetName())
	}

	return framework.NewStatus(framework.Success, "")
}

func (y *Yoda) ScoreExtensions() framework.ScoreExtensions {
	return y
}

func (y *Yoda) PreBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	if nodeInfo, err := y.handle.SnapshotSharedLister().NodeInfos().Get(nodeName); err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("prebind get node info error: %+v", nodeName))
	} else {
		klog.V(3).Infof("prebind node info : %+v", nodeInfo.Node())
		return framework.NewStatus(framework.Success, "")
	}
}

func (y *Yoda) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}
