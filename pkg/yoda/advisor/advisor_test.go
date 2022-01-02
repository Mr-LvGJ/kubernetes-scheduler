package advisor

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
)

var Resources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory, v1.ResourcePods, v1.ResourceStorage, v1.ResourceEphemeralStorage}

type resourceToWeightMap map[v1.ResourceName]int64

func TestInfo(t *testing.T) {
	var res Result

	info, err := res.Init()
	if err != nil {
		t.Errorf("failed to get info , err: %v", err)
	}
	for i := range info.Info {
		fmt.Println(info.Info[i])
	}
}

func TestResources(t *testing.T) {
	tr := make(resourceToWeightMap)
	for _, resourse := range Resources {
		fmt.Println("sss")

		tr[resourse] = 1
	}
	fmt.Println("sss")
}
