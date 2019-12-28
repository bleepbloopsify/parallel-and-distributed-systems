package mapreduce

import (
	"fmt"
)

type WorkerInfo struct {
	address string
}

// KillWorkers will clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
func (mr *MapReduce) KillWorkers() []int {
	l := make([]int, 0)
	for _, w := range mr.Workers {
		DPrintf("DoWork: shutdown %s\n", w.address)
		args := &ShutdownArgs{}
		var reply ShutdownReply
		ok := call(w.address, "Worker.Shutdown", args, &reply)
		if ok == false || reply.OK == false {
			fmt.Printf("DoWork: RPC %s shutdown error\n", w.address)
		} else {
			l = append(l, reply.Njobs)
		}
	}

	return l
}

// RunMaster starts up the MapReduce process
func (mr *MapReduce) RunMaster() []int {

	mapJobs := make(chan bool, mr.nMap)

	for i := 0; i < mr.nMap; i++ {
		go func(idx int) {
			worker := <-mr.registerChannel

			var reply DoJobReply
			args := &DoJobArgs{mr.file, Map, idx, mr.nReduce}
			call(worker, "Worker.DoJob", args, &reply)
			mr.registerChannel <- worker
			mapJobs <- true
		}(i)
	}

	for i := 0; i < mr.nMap; i++ {
		<-mapJobs
	}
	close(mapJobs)

	reduceJobs := make(chan bool, mr.nReduce)
	for i := 0; i < mr.nReduce; i++ {
		go func(idx int) {
			worker := <-mr.registerChannel

			var reply DoJobReply
			args := &DoJobArgs{mr.file, Reduce, idx, mr.nMap}
			call(worker, "Worker.DoJob", args, &reply)

			mr.registerChannel <- worker
			reduceJobs <- true
		}(i)
	}

	for i := 0; i < mr.nReduce; i++ {
		<-reduceJobs
	}
	close(reduceJobs)

	return mr.KillWorkers()
}
