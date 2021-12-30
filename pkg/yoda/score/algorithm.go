package score

import (
	"errors"

	"github.com/Mr-LvGJ/Yoda-Scheduler/pkg/yoda/advisor"
	"k8s.io/klog/v2"

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

func CalculateScore(s *advisor.Result, state *framework.CycleState, pod *v1.Pod, info *framework.NodeInfo) (uint64, error) {
	d, err := state.Read("nodeInfo")
	if err != nil {
		return 0, errors.New("Error Get CycleState Info Max Error: " + err.Error())
	}
	data, ok := d.(*advisor.Result)
	if !ok {
		return 0, errors.New("The Type is not result ")
	}
	return CalculateBasicScore2(data.Info, s, pod, info)

	//return CalculateBasicScore(data.Value, s, pod) + CalculateAllocateScore(info, s) + CalculateActualScore(s), nil
}

func CalculateBasicScore2(info map[string]*advisor.NodeInfo, s *advisor.Result, pod *v1.Pod, nodeInfo *framework.NodeInfo) (uint64, error) {
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
