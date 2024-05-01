package modules

import (
	"context"
	"fmt"

	internalapi "k8s.io/cri-api/pkg/apis"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func UpdateContainerResources(client internalapi.RuntimeService, id string, resource *pb.ContainerResources) {

	err := client.UpdateContainerResources(context.TODO(), id, resource)
	if err != nil {
		fmt.Println(err)
	}
}

func MonitoringPodResources(client internalapi.RuntimeService) ([]PodData, []*pb.ContainerResources) {
	var containerResourceSet = make([]*pb.ContainerResources, 0) //Dynamic array to store container system metric

	// 우선 call by reference 방식이 아닌 call by value 방식으로 구현 및 작동 확인함. 추후 공부 후 call by reference 방식으로 변경 필요
	podInfoSet := PodInfoInit()
	// get pod stats
	podInfoSet = GetPodStatsInfo(client, podInfoSet)

	// get container stats
	podInfoSet, containerResourceSet = GetContainerStatsInfo(client, podInfoSet, containerResourceSet)

	// get memory usage percents each containers
	podInfoSet = GetmemoryUsagePercents(podInfoSet)

	return podInfoSet, containerResourceSet
}

func RemoveContainer(client internalapi.RuntimeService, selectContainerId []string, selectContainerResource []*pb.ContainerResources) ([]string, []*pb.ContainerResources) {
	err := client.RemoveContainer(context.TODO(), selectContainerId[len(selectContainerId)-1])
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println()
	fmt.Println("Remove Container Id:", selectContainerId[len(selectContainerId)-1])

	return selectContainerId[:len(selectContainerId)-1], selectContainerResource[:len(selectContainerResource)-1]
}
