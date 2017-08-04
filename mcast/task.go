package mcast

import "fmt"

type Task struct {
	TaskId uint8
	RTime  float32
	PType  uint8
	Pid    uint8
}

func NewTask(taskId uint8, rtime float32, ptype, pid uint8) *Task {
	return &Task{TaskId: taskId, RTime: rtime, PType: ptype, Pid: pid}
}

func (t *Task) Dump(sp string) {
	rtype := "REQ"
	del := ""
	if t.PType == ACK {
		rtype = "ACK"
		//		del = "-"
		del = fmt.Sprintf("-%d", t.Pid)
	}
	s := fmt.Sprintf("%s%d%s: %1.1f", rtype, t.TaskId, del, t.RTime)
	fmt.Println(sp + s)

}
