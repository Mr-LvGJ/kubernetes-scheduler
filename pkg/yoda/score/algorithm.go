package score

import (
	"context"
	"errors"
	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/filter"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"

	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/advisor"
	"k8s.io/klog/v2"
	schedutil "k8s.io/kubernetes/pkg/scheduler/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Sum is from collection/collection.go
// var Sum = []string{"Cores","FreeMemory","Bandwidth","MemoryClock","MemorySum","Number","Memory"}

const (
	BandwidthWeight   = 1
	ClockWeight       = 1
	CoreWeight        = 2
	PowerWeight       = 1
	FreeMemoryWeight  = 3
	TotalMemoryWeight = 1
	ActualWeight      = 2
	DiskIOWeight      = 100

	AllocateWeight = 3
)

type RequestInfo struct {
	cpuTotal    int64
	memoryTotal int64
}

type resourceScorer struct {
	resToWeightmap resourceToWeightMap
}
type resourceToWeightMap map[v1.ResourceName]int64

func CalculateScore(s *advisor.Result, state *framework.CycleState, pod *v1.Pod, info *framework.NodeInfo, nodeList []*framework.NodeInfo, client *redis.Client) (uint64, error) {
	d, err := state.Read("nodeInfo")
	if err != nil {
		return 0, errors.New("Error Get CycleState Info Max Error: " + err.Error())
	}
	data, ok := d.(*advisor.Result)
	if !ok {
		return 0, errors.New("The Type is not result ")
	}
	ctx := context.Background()
	result, err := client.Get(ctx, "S-"+info.Node().GetName()).Result()
	klog.V(3).Infof("!!!!!!!!!!!!!!!res: %v", result)

	if err != redis.Nil{
		klog.V(3).Infof("exist!!!!===== score: %v", result)
		return filter.StrToUint64(result),nil
	}

	klog.V(4).Info("============== begin M-tmp ====================")

	var u_avg float64
	for _, node := range nodeList {
		name := node.Node().GetName()
		Di := data.Info[name].DiskIO
		Ui := Di / 50.0
		u_avg = u_avg + Ui
		client.Set(ctx, "U-"+name, Ui, 0)
	}
	u_avg = u_avg / float64(len(nodeList))
	client.Set(ctx, "U-AVG", u_avg, 0)
	client.Set(ctx, "nodeLen", len(nodeList), 0)
	var M_tmp float64
	for _, node := range nodeList {
		name := node.Node().GetName()
		Ui, _ := client.Get(ctx, "U-"+name).Float64()
		tmp := (Ui - u_avg) * (Ui - u_avg)
		M_tmp = M_tmp + tmp
	}
	M_tmp = M_tmp / float64(len(nodeList))
	client.Set(ctx, "M-tmp", M_tmp, 0)
	klog.V(4).Info("============== end M-tmp ====================")

	return BalancedDiskIOPriority(data.Info, pod, info, client, nodeList)

	//return CalculateBasicScore2(data.Info, s, pod, info, client)

	//return CalculateBasicScore(data.Value, s, pod) + CalculateAllocateScore(info, s) + CalculateActualScore(s), nil
}

func BalancedDiskIOPriority(info map[string]*advisor.NodeInfo, pod *v1.Pod, nodeInfo2 *framework.NodeInfo, client *redis.Client, nodeList []*framework.NodeInfo)(uint64, error){
	M_max := 0.0
	M_min := 1000000.0
	nameToM := make(map[string]float64)
	var res uint64
	for _, nodeInfo := range nodeList{
		klog.V(4).Info("============== begin calculate ====================")
		klog.V(4).Info("============== begin calculate ====================")
		klog.V(4).Info("============== begin calculate ====================")
		klog.V(4).Info("============== begin calculate ====================")
		name := nodeInfo.Node().GetName()
		klog.V(4).Infof("name：%v", name)
		currentInfo := info[name]
		Di := currentInfo.DiskIO
		klog.V(4).Infof("Di：%v", Di)
		Rio, _ := strconv.ParseFloat(pod.Annotations["diskIO"], 32)
		klog.V(4).Infof("Rio：%v", Rio)
		Tj := Di + Rio
		Fj := Tj / 100.0
		klog.V(4).Infof("Tj：%v, Fj: %v", Tj, Fj)
		Uj, _ := client.Get(context.Background(), "U-"+nodeInfo.Node().GetName()).Float64()
		klog.V(4).Infof("Uj：%v", Uj)
		U_avg, _ := client.Get(context.Background(), "U-AVG").Float64()
		klog.V(4).Infof("U_avg：%v", U_avg)
		len, _ := client.Get(context.Background(),"nodeLen").Float64()
		F_avg := U_avg - (Uj - Fj) /len
		klog.V(4).Infof("F_avg：%v", F_avg)
		M_tmp, _ := client.Get(context.Background(), "M-tmp").Float64()
		klog.V(4).Infof("M_tmp：%v", M_tmp)
		Mj := M_tmp - ((Uj - U_avg)*(Uj - U_avg) - (Fj - F_avg)*(Fj - F_avg))/ len
		klog.V(4).Infof("Mj：%v", Mj)
		if Mj > M_max{
			M_max = Mj
		}
		if (Mj < M_min){
			M_min = Mj
		}
		nameToM[name] = Mj
		klog.V(4).Info("============== end calculate ====================")
		klog.V(4).Info("============== end calculate ====================")
		klog.V(4).Info("============== end calculate ====================")
	}
	for _, nodeInfo := range nodeList {
		name := nodeInfo.Node().GetName()
		Mi := nameToM[name]
		Si := 100.0 - (100.0 * (Mi - M_min) / (M_max - M_min))
		klog.V(4).Infof("Mi：%v, M_min: %v, M_max: %v, Si: %v", Mi, M_min,M_max,Si)
		if name == nodeInfo2.Node().GetName() {
			res = uint64(Si)
		}
		client.Set(context.Background(), "S-"+name, Si, 60*time.Minute)
	}


	klog.V(3).Infof("===== score: %v", res)
	return res, nil
}

func CalculateBasicScore2(info map[string]*advisor.NodeInfo, s *advisor.Result, pod *v1.Pod, nodeInfo *framework.NodeInfo, client *redis.Client) (uint64, error) {
	var basicScore int64

	requestInfo := getCpuRequest(pod)
	capacityInfo := nodeInfo.Allocatable.Clone()

	klog.V(3).Infof("This Node allocatable: %+v, %+v", capacityInfo.Memory, capacityInfo.MilliCPU)
	klog.V(3).Infof("This pod total request cpu: %v, memory: %v", requestInfo.cpuTotal, requestInfo.memoryTotal)
	request_diskIO := pod.Annotations["diskIO"]
	currentInfo := info[nodeInfo.Node().GetName()]

	diskIOScore := DiskIOWeight * (100 - int64(currentInfo.DiskIO))
	cpuScore := CoreWeight * (100 - currentInfo.Cpu)
	memoryScore := FreeMemoryWeight * (100 - currentInfo.Memory)

	basicScore = diskIOScore + int64(cpuScore) + int64(memoryScore)

	klog.V(3).Infof("request_diskIO: %v, currentInfo diskIO: %v", request_diskIO, currentInfo.DiskIO)
	klog.V(3).Infof("===== score: %v", basicScore)
	return uint64(basicScore), nil
}

func getCpuRequest(pod *v1.Pod) *RequestInfo {
	res := &RequestInfo{}
	containers := pod.Spec.Containers
	for _, container := range containers {
		res.cpuTotal = res.cpuTotal + container.Resources.Requests.Cpu().Value()
		res.memoryTotal += container.Resources.Requests.Memory().Value()
	}
	return res
}
func CalculateResourceAllocatableRequest(nodeInfo *framework.NodeInfo, pod *v1.Pod, resource v1.ResourceName, enablePodOverhead bool) (int64, int64) {
	podRequest := CalculatePodResourceRequest(pod, resource, enablePodOverhead)
	// If it's an extended resource, and the pod doesn't request it. We return (0, 0)
	// as an implication to bypass scoring on this resource.
	if podRequest == 0 && schedutil.IsScalarResourceName(resource) {
		return 0, 0
	}

	switch resource {
	case v1.ResourceCPU:
		return nodeInfo.Allocatable.MilliCPU, (nodeInfo.NonZeroRequested.MilliCPU + podRequest)
	case v1.ResourceMemory:
		return nodeInfo.Allocatable.Memory, (nodeInfo.NonZeroRequested.Memory + podRequest)
	case v1.ResourceEphemeralStorage:
		return nodeInfo.Allocatable.EphemeralStorage, (nodeInfo.Requested.EphemeralStorage + podRequest)
	default:
		if _, exists := nodeInfo.Allocatable.ScalarResources[resource]; exists {
			return nodeInfo.Allocatable.ScalarResources[resource], (nodeInfo.Requested.ScalarResources[resource] + podRequest)
		}
	}
	if klog.V(10).Enabled() {
		klog.Infof("requested resource %v not considered for node score calculation", resource)
	}
	return 0, 0
}

// calculatePodResourceRequest returns the total non-zero requests. If Overhead is defined for the pod and the
// PodOverhead feature is enabled, the Overhead is added to the result
// podResourceRequest = max(sum(podSpec.Containers), podSpec.InitContainers) + overHead
func CalculatePodResourceRequest(pod *v1.Pod, resource v1.ResourceName, enablePodOverhead bool) int64 {
	var podRequest int64
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &container.Resources.Requests)
		podRequest += value
	}

	for i := range pod.Spec.InitContainers {
		initContainer := &pod.Spec.InitContainers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &initContainer.Resources.Requests)
		if podRequest < value {
			podRequest = value
		}
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && enablePodOverhead {
		if quantity, found := pod.Spec.Overhead[resource]; found {
			podRequest += quantity.Value()
		}
	}

	return podRequest
}

//func CalculateBasicScore(value collection.MaxValue, scv *advisor.Result, pod *v1.Pod) uint64 {
//	var cardScore uint64
//	if ok, number := filter.PodFitsNumber(pod, scv); ok {
//		isFitsMemory, memory := filter.PodFitsMemory(number, pod, scv)
//		isFitsClock, clock := filter.PodFitsClock(number, pod, scv)
//		if isFitsClock && isFitsMemory {
//			for _, card := range scv.Status.CardList {
//				if card.FreeMemory >= memory && card.Clock >= clock {
//					cardScore += CalculateCardScore(value, card)
//				}
//			}
//		}
//	}
//	return cardScore
//}
//
//func CalculateCardScore(value collection.MaxValue, card scv.Card) uint64 {
//	var (
//		bandwidth   = card.Bandwidth * 100 / value.MaxBandwidth
//		clock       = card.Clock * 100 / value.MaxBandwidth
//		core        = card.Core * 100 / value.MaxCore
//		power       = card.Power * 100 / value.MaxPower
//		freeMemory  = card.FreeMemory * 100 / value.MaxFreeMemory
//		totalMemory = card.TotalMemory * 100 / value.MaxTotalMemory
//	)
//	return uint64(bandwidth*BandwidthWeight+clock*ClockWeight+core*CoreWeight+power*PowerWeight) +
//		freeMemory*FreeMemoryWeight + totalMemory*TotalMemoryWeight
//}
//
//func CalculateActualScore(scv *advisor.Result) uint64 {
//	return (scv.Status.FreeMemorySum * 100 / scv.Status.TotalMemorySum) * ActualWeight
//}
//
//func CalculateAllocateScore(info *framework.NodeInfo, scv *advisor.Result) uint64 {
//	allocateMemorySum := uint64(0)
//	for _, pod := range info.Pods {
//		if mem, ok := pod.Pod.GetLabels()["scv/memory"]; ok {
//			allocateMemorySum += filter.StrToUint64(mem)
//		}
//	}
//
//	if scv.Status.TotalMemorySum < allocateMemorySum {
//		return 0
//	}
//
//	return (scv.Status.TotalMemorySum - allocateMemorySum) * 100 / scv.Status.TotalMemorySum * AllocateWeight
//}
