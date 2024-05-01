package main

import (
	"fmt"
	"time"

	"k8s.io/kubernetes/pkg/kubelet/cri/remote"

	mod "iitp/program/modules"
)

func IsMobilityPodRunning(podInfoSet []mod.PodData) bool {

	for i := 0; i < len(podInfoSet); i++ {
		if podInfoSet[i].Namespace != "default" {
			// fmt.Println("Mobility Pod is Running!")
			return true
		}
	}
	return false
}

func main() {
	const ENDPOINT string = "unix:///var/run/containerd/containerd.sock"
	const DEFAULT_CPU_QUOTA int64 = 200000

	/*
		// kubernetes api 클라이언트 생성하는 모듈
		clientset := mod.InitClient()
		if clientset != nil {
			fmt.Println("123")
		}*/

	//get new internal client service
	client, err := remote.NewRemoteRuntimeService(ENDPOINT, time.Second*2, nil)
	if err != nil {
		panic(err)
	}
	// remote.NewRemoteImageService("unix:///var/run/containerd/containerd.sock", time.Second*2, nil)

	// loop
	var loop = 0
	var isRestricted bool = false
	temp := make(map[string][]uint64)

	for {
		// Monitoring Pod Resources
		podInfoSet, containerResourceSet := mod.MonitoringPodResources(client)

		if IsMobilityPodRunning(podInfoSet) {
			if !isRestricted {
				for i := 0; i < len(podInfoSet); i++ {
					if podInfoSet[i].Namespace != "default" {
						continue
					}
					containerResourceSet[i].Linux.CpuQuota = mod.LIMIT_CPU_QUOTA // limit cpu usage to 10m
					mod.UpdateContainerResources(client, podInfoSet[i].ContainerData.Id, containerResourceSet[i])
				}
				fmt.Println("Restrict is completed")
				isRestricted = true
			}
		} else {
			if isRestricted {
				for i := 0; i < len(podInfoSet); i++ {
					if podInfoSet[i].Namespace != "default" {
						continue
					}
					containerResourceSet[i].Linux.CpuQuota = mod.DEFAULT_CPU_QUOTA // limit cpu usage to 10m
					mod.UpdateContainerResources(client, podInfoSet[i].ContainerData.Id, containerResourceSet[i])
				}
				fmt.Println("Release is completed")
				isRestricted = false
			}
		}

		if loop%5 == 0 {
			fmt.Printf("%-10s\t", "NAME")
			fmt.Printf("%-12s\t", "CPU USAGE")
			fmt.Printf("%-5s\t", "MEMORY LIMIT")
			fmt.Printf("%-5s\n", "MEMORY USAGE")
		}

		for i := 0; i < len(podInfoSet); i++ {
			val, exists := temp[podInfoSet[i].Name]
			if !exists {
				val = make([]uint64, 0)
				val = append(val, 0)
				temp[podInfoSet[i].Name] = val
			}
			val = append(val, podInfoSet[i].ContainerData.ResourceData.CpuUsageCoreNanoSeconds)
			temp[podInfoSet[i].Name] = val

			if loop%5 == 0 {
				fmt.Printf("%-10.10s\t", podInfoSet[i].Name)
				fmt.Printf("%8dm\t", (podInfoSet[i].ContainerData.ResourceData.CpuUsageCoreNanoSeconds-val[len(val)-2])/1000000)
				fmt.Printf("%4.2f GiB\t", float64(podInfoSet[i].ContainerData.LinuxResourceData.MemoryLimitInBytes)/1073741824.0)
				fmt.Printf("%4.2f%%\n", podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents*100.0)
			}

			// 모빌리티 컨테이너는 관리대상 아님
			if podInfoSet[i].Namespace != "default" {
				continue
			}
			// 너무 메모리 사용량이 작은 컨테이너는 관리대상 아님
			if podInfoSet[i].ContainerData.ResourceData.MemoryUsageBytes < 838860800 {
				continue
			}

			if containerResourceSet[i].Linux.MemoryLimitInBytes > 10000000000 {
				var alterMem int64 = 1048576000
				containerResourceSet[i].Linux.MemoryLimitInBytes = alterMem
				mod.UpdateContainerResources(client, podInfoSet[i].ContainerData.Id, containerResourceSet[i])
			}

			if podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents > 0.85 {
				var alterMem float64 = float64(containerResourceSet[i].Linux.MemoryLimitInBytes) * 1.1
				containerResourceSet[i].Linux.MemoryLimitInBytes = int64(alterMem)
				mod.UpdateContainerResources(client, podInfoSet[i].ContainerData.Id, containerResourceSet[i])
			}

			if podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents < 0.75 {
				var alterMem float64 = float64(containerResourceSet[i].Linux.MemoryLimitInBytes) * 0.9
				// 기본 컨테이너 사이즈가 1GiB보다 작아지지 않게
				if alterMem < 1073741824 {
					alterMem = 1073741824
				}
				containerResourceSet[i].Linux.MemoryLimitInBytes = int64(alterMem)
				mod.UpdateContainerResources(client, podInfoSet[i].ContainerData.Id, containerResourceSet[i])
			}
		}
		if loop%5 == 0 {
			fmt.Println("---------------------------------------------")
		}
		loop++
		time.Sleep(250 * time.Millisecond)
	}
}
