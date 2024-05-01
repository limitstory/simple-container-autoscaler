package modules

func GetmemoryUsagePercents(podInfoSet []PodData) []PodData {
	// get current container memory usage and limit value
	for i := 0; i < len(podInfoSet); i++ {
		containerMemoryUsages := podInfoSet[i].ContainerData.ResourceData.MemoryUsageBytes
		// if limit is not set, it will appear as 0; if set, it will output normally.
		containerMemoryLimits := podInfoSet[i].ContainerData.LinuxResourceData.MemoryLimitInBytes

		// exception handling
		// container without limit set, not burstable container
		if containerMemoryLimits == 0 {
			podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents = 0
		}
		podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents = float64(containerMemoryUsages) / float64(containerMemoryLimits)
	}

	return podInfoSet
}

func RemovePodofPodInfoSet(podInfoSet []PodData, i int) []PodData {
	podInfoSet[i] = podInfoSet[len(podInfoSet)-1]
	return podInfoSet[:len(podInfoSet)-1]
}
