package mcast

import (
	"container/list"
	"fmt"
	"os"
	"sync"
	"time"
)

type Process struct {
	pid       uint8
	tick      uint8
	rchan     <-chan Packet
	commQueue *list.List
	//	mutex     *sync.Mutex
	taskList *list.List
	sp       string
	sync.Mutex
	schan chan<- Packet
}

func NewProcess(pid, tick uint8, rchan <-chan Packet, sp string) (*Process, <-chan Packet) {
	mychan := make(chan Packet)

	var proc Process
	proc.pid = pid
	proc.tick = tick
	proc.rchan = rchan
	proc.commQueue = list.New()
	proc.taskList = list.New()
	//	proc.mutex = new(sync.Mutex)
	proc.sp = sp
	proc.schan = mychan
	return &proc, mychan
}

func (proc *Process) incTick() {
	//proc.mutex.Lock()
	proc.Lock()
	proc.tick++
	proc.Unlock()
	//proc.mutex.Unlock()
}

func (proc *Process) trun() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			proc.incTick()
		}
	}()
}

func (proc *Process) getTime() float32 {
	return float32(proc.tick) + float32(proc.pid)*0.1
}

func (proc *Process) adjustTick(time float32) {
	if time > float32(proc.tick) {
		//		proc.mutex.Lock()
		proc.Lock()
		proc.tick = uint8(time + 1)
		proc.Unlock()
		//		proc.mutex.Unlock()
	}
}
func (proc *Process) createAndAddTask(packet *Packet) {
	t := NewTask(packet.GetReqId(), packet.GetTime(), packet.GetType(), packet.GetSender())
	proc.adjustTick(packet.GetTime())
	proc.addTask(t)
	proc.showTasks(proc.taskList)

}

func (proc *Process) Run() {
	proc.trun() // timer
	go func() {
		for {
			proc.incTick()
			//			time.Sleep(1 * time.Second)
			//			fmt.Printf("%srun in process in proc %d\n", proc.sp, proc.pid)
			select {
			case packet := <-proc.rchan:
				{
					switch packet.GetType() {
					case SND:
						{ //multicast packet send
							fmt.Printf("%smulticast send by proc %d\n", proc.sp, proc.pid)
							rp := NewPacket(REQ, proc.pid, proc.getTime(), proc.pid)
							proc.schan <- *rp
						}
					case REQ:
						{

							proc.createAndAddTask(&packet)
							proc.incTick()
							ctime := proc.getTime()
							fmt.Printf("%ssend ack in proc %d: %1.1f\n", proc.sp, proc.pid, ctime)
							rp := NewPacket(ACK, proc.pid, ctime, packet.GetReqId())
							proc.schan <- *rp
						}
					case ACK:
						{
							fmt.Printf("%sack arrived in proc %d\n", proc.sp, proc.pid)
							proc.createAndAddTask(&packet)
						}
					default:
						{
							os.Exit(1)
						}
					}
				}
			}
			proc.execute()
		}
	}()
}

func filter(l *list.List, len int, f func(e *list.Element) (interface{}, bool)) *list.List {
	if len == 0 {
		return list.New()
	}
	car := l.Front()
	l.Remove(car)
	v, ok := f(car)
	nl := filter(l, l.Len(), f)
	if ok {
		nl.PushFront(v)
	}
	return nl
}

func (proc *Process) copy(l *list.List) *list.List {
	nl := list.New()
	for e := l.Front(); e != nil; e = e.Next() {
		nl.PushBack(e.Value)
	}
	return nl
}

func (proc *Process) filterRequest(l *list.List) *list.List {
	l2 := proc.copy(l)
	nl := filter(l2, l2.Len(), func(e *list.Element) (interface{}, bool) {
		t, ok := e.Value.(*Task)
		if ok {
			if t.PType == REQ {
				return t, true
			}
		}
		return nil, false
	})
	return nl
}

func (proc *Process) filterByTaskId(l *list.List, taskId uint8) *list.List {
	l2 := proc.copy(l)
	nl := filter(l2, l2.Len(), func(e *list.Element) (interface{}, bool) {
		t, ok := e.Value.(*Task)
		if ok {
			if t.TaskId == taskId {
				return t, true
			}
		}
		return nil, false
	})
	return nl
}

func (proc *Process) getTargetTask(reqList *list.List) uint8 {
	var targetTid uint8
	var time float32 = 1000000
	for e := reqList.Front(); e != nil; e = e.Next() {
		c, ok := e.Value.(*Task)
		if ok {
			//				c.Dump(proc.sp)
			//				fmt.Printf("%staskId = %d\n", proc.sp, c.TaskId)
			tl := proc.filterByTaskId(proc.taskList, c.TaskId)
			//				proc.showTasks(tl)
			te := tl.Back()
			tsk, ok2 := te.Value.(*Task)
			if ok2 {
				if tsk.RTime < time {
					time = tsk.RTime
					targetTid = tsk.TaskId
				}
			}
		} else {
			panic("nononon")
		}
	}
	//	fmt.Printf("%sexecute target %d:%1.1f\n", proc.sp, targetTid, time)
	return targetTid
}

func (proc *Process) execute() {
	if proc.taskList.Len() == 0 {
		fmt.Printf("%sno task\n", proc.sp)
		return
	}
	reqList := proc.filterRequest(proc.taskList)
	if proc.taskList.Len() == reqList.Len()*3 {
		targetTid := proc.getTargetTask(reqList)
		fmt.Printf("%sexecute task %d\n", proc.sp, targetTid)
		//		fmt.Printf("%srest of tasks are:\n", proc.sp)
		proc.taskList = filter(proc.taskList, proc.taskList.Len(), func(e *list.Element) (interface{}, bool) {
			t, ok := e.Value.(*Task)
			if ok {
				if t.TaskId != targetTid {
					return t, true
				}
			}
			return nil, false
		})
		//		proc.showTasks(proc.taskList)
		proc.execute()

	}
}

func (proc *Process) ShowProc() {
	fmt.Printf("pid = %d, tick = %d\n", proc.pid, proc.tick)
}

func (proc *Process) addTask(task *Task) {
	var target *list.Element
	for e := proc.taskList.Front(); e != nil; e = e.Next() {
		c, ok := e.Value.(*Task)
		if ok {
			if c.RTime > task.RTime {
				target = e
				break
			}
		} else {
			panic("noooooooooo")
		}
	}
	if target == nil {
		proc.taskList.PushBack(task)
	} else {
		proc.taskList.InsertBefore(task, target)
	}

	//	proc.taskList.PushBack(task)
}

func (proc *Process) showTasks(l *list.List) {
	for e := l.Front(); e != nil; e = e.Next() {
		c, ok := e.Value.(*Task)
		if ok {
			c.Dump(proc.sp)
		} else {
			panic("nononon")
		}
	}
}

func (proc *Process) AddCommand(comm *Command) {
	proc.commQueue.PushBack(comm)
}

func (proc *Process) GetCommand() *Command {
	c := proc.commQueue.Remove(proc.commQueue.Front())
	v, ok := c.(*Command)
	if ok {
		return v
	} else {
		panic("element should be command")
	}
}
