package advisor

import (
"encoding/json"
"fmt"
"io/ioutil"
"net/http"
"net/url"
"strconv"
"strings"
)

const host string = "110.40.247.174:32500"
const cpuQueryL string = "sum by (kubernetes_io_hostname, instance)(rate(container_cpu_usage_seconds_total{image!=\"\",pod!=\"\"}[1m]) * 100)"
const memoryQueryL string = "(node_memory_MemTotal_bytes-node_memory_MemFree_bytes-node_memory_Buffers_bytes-node_memory_Cached_bytes)/node_memory_MemTotal_bytes{kubernetes_io_hostname!=\"\"} * 100"
const diskIOQueryL string = "( rate(node_disk_read_bytes_total{device=\"vda\"}[1m]) + rate(node_disk_written_bytes_total{device=\"vda\"}[1m])) /1024 /1024"
const networkIOUpQueryL string = "sum by (kubernetes_io_hostname,instance) (rate (node_network_transmit_bytes_total{kubernetes_io_hostname!=\"\"}[1m])) /1024/1024"
const networkIODownQueryL string = "sum by (kubernetes_io_hostname,instance)(rate (node_network_receive_bytes_total{kubernetes_io_hostname!=\"\"}[1m]))/ 1024 /1024"

type Result struct {
	Info map[string]*NodeInfo
}

type NodeInfo struct {
	Cpu           float64
	Memory        float64
	DiskIO        float64
	NetworkIOUp   float64
	NetworkIODown float64
}

type PrometheusResp struct {
	Status string `json:"status"`
	Data   PrometheusData
}

type PrometheusData struct {
	ResultType string `json:"resultType"`
	Result     []PrometheusResult
}

type PrometheusResult struct {
	Metric PrometheusMetric
	Value  []interface{}
}

type PrometheusMetric struct {
	Name                 string `json:"__name__"`
	Job                  string `json:"job"`
	Instance             string `json:"instance"`
	KubernetesIoHostname string `json:"kubernetes_io_hostname"`
}

func getCpu() ([]PrometheusResult, error) {
	urlValues := url.Values{}
	urlValues.Add("query", cpuQueryL)
	// urlValues.Add("query", "up")
	var ret []PrometheusResult
	resp, err := http.PostForm("http://"+host+"/api/v1/query", urlValues)
	if err != nil {
		return ret, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	// fmt.Println(string(body))

	var res PrometheusResp
	_ = json.Unmarshal(body, &res)
	ret = res.Data.Result
	return ret, nil
}

func getMemory() ([]PrometheusResult, error) {
	urlValues := url.Values{}
	urlValues.Add("query", memoryQueryL)
	// urlValues.Add("query", "up")
	var ret []PrometheusResult
	resp, err := http.PostForm("http://"+host+"/api/v1/query", urlValues)
	if err != nil {
		return ret, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	// fmt.Println()
	// fmt.Println("==" + string(body))

	var res PrometheusResp
	_ = json.Unmarshal(body, &res)
	ret = res.Data.Result
	return ret, nil
}

func getDiskIO() ([]PrometheusResult, error) {
	urlValues := url.Values{}
	urlValues.Add("query", diskIOQueryL)
	// urlValues.Add("query", "up")
	var ret []PrometheusResult
	resp, err := http.PostForm("http://"+host+"/api/v1/query", urlValues)
	if err != nil {
		return ret, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	// fmt.Println(string(body))

	var res PrometheusResp
	_ = json.Unmarshal(body, &res)
	ret = res.Data.Result
	return ret, nil
}

func getNetworkIOQueryL(netInfo string) ([]PrometheusResult, error) {
	urlValues := url.Values{}
	urlValues.Add("query", netInfo)
	// urlValues.Add("query", "up")
	var ret []PrometheusResult
	resp, err := http.PostForm("http://"+host+"/api/v1/query", urlValues)
	if err != nil {
		return ret, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	// fmt.Println(string(body))

	var res PrometheusResp
	_ = json.Unmarshal(body, &res)
	ret = res.Data.Result
	return ret, nil
}

func (r Result) Init() (Result, error) {
	cpu, err := getCpu()
	var res Result
	if err != nil {
		return res, err
	}
	res.Info = make(map[string]*NodeInfo)
	for i := 0; i < len(cpu); i++ {
		name := strings.Trim(cpu[i].Metric.KubernetesIoHostname, " ")
		// fmt.Printf("%v", name)
		_, ok := res.Info[name]
		if ok {
			fmt.Println("Error! cpu info no exist!")
		} else {
			tmpCpu, err := strconv.ParseFloat(cpu[i].Value[1].(string), 64)

			if err != nil {
				return res, err
			}
			res.Info[name] = &NodeInfo{Cpu: tmpCpu}
			// fmt.Println()
			// fmt.Printf("---------%#v", res.Info[name])
			// res.Info[name] = nodeInfo
		}
		// fmt.Println(nodeInfo.Cpu)
	}

	memory, err := getMemory()
	if err != nil {
		return res, err
	}
	for i := 0; i < len(memory); i++ {
		name := strings.Trim(memory[i].Metric.KubernetesIoHostname, " ")
		_, ok := res.Info[name]
		if ok {
			tmpMemory, err := strconv.ParseFloat(memory[i].Value[1].(string), 64)
			if err != nil {
				return res, err
			}
			res.Info[name].Memory = tmpMemory
		} else {
			fmt.Println("No memory info!")
		}
	}

	diskIO, err := getDiskIO()
	if err != nil {
		return res, err
	}
	for i := 0; i < len(diskIO); i++ {
		name := strings.Trim(diskIO[i].Metric.KubernetesIoHostname, " ")
		if name == "" {
			name = strings.Trim(diskIO[i].Metric.Instance, " ")
		}
		// fmt.Printf("%v", name)

		_, ok := res.Info[name]
		if ok {
			tmpDiskIO, err := strconv.ParseFloat(diskIO[i].Value[1].(string), 64)
			if err != nil {
				return res, err
			}
			res.Info[name].DiskIO = tmpDiskIO
		} else {
			fmt.Printf("No Disk Info: %v \n", name)
		}
	}

	networkIOUp, err := getNetworkIOQueryL(networkIOUpQueryL)
	if err != nil {
		return res, nil
	}

	for i := 0; i < len(networkIOUp); i++ {
		name := strings.Trim(networkIOUp[i].Metric.KubernetesIoHostname, " ")
		if name == "" {
			name = strings.Trim(networkIOUp[i].Metric.Instance, " ")
		}

		_, ok := res.Info[name]
		if ok {
			tmpNetworkIOup, err := strconv.ParseFloat(networkIOUp[i].Value[1].(string), 64)
			if err != nil {
				return res, nil
			}
			res.Info[name].NetworkIOUp = tmpNetworkIOup
		} else {
			fmt.Printf("No network transimit info: %v \n", name)
		}
	}

	networkIODown, err := getNetworkIOQueryL(networkIODownQueryL)
	if err != nil {
		return res, nil
	}
	for i := 0; i < len(networkIODown); i++ {
		name := strings.Trim(networkIODown[i].Metric.KubernetesIoHostname, " ")
		if name == "" {
			name = strings.Trim(networkIODown[i].Metric.Instance, " ")
		}

		_, ok := res.Info[name]
		if ok {
			tmpNetworkIODown, err := strconv.ParseFloat(networkIODown[i].Value[1].(string), 64)
			if err != nil {
				return res, nil
			}
			res.Info[name].NetworkIODown = tmpNetworkIODown
		} else {
			fmt.Printf("No network recevice info: %v \n", name)
		}
	}
	// for ss := range res.Info {
	// 	fmt.Println(ss, "info:", res.Info[ss])
	// }
	return res, nil
}

// func main() {
// 	var r Result
// 	r.init()
// }
