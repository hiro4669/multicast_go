package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"./mcast"
)

const delay = 4

func run(mcan <-chan string) {
	pchan1 := make(chan mcast.Packet)
	pchan2 := make(chan mcast.Packet)
	sp := "                                   "

	proc1, rchan1 := mcast.NewProcess(1, 1, pchan1, "")
	proc2, rchan2 := mcast.NewProcess(2, 2, pchan2, sp)
	proc1.Run()
	proc2.Run()
	for {
		time.Sleep(1 * time.Second)
		log.Println("working...")
		select {
		case command := <-mcan:
			switch command {
			case "exit":
				{
					fmt.Println("break")
					return
				}
			case "send":
				{
					packet1 := mcast.NewPacket(mcast.SND, 1, 10.0, 1)
					pchan1 <- *packet1
					time.Sleep(500 * time.Millisecond)
					packet2 := mcast.NewPacket(mcast.SND, 2, 10.0, 1)
					pchan2 <- *packet2
				}
			case "send1":
				{
					packet := mcast.NewPacket(mcast.SND, 1, 10.0, 1)
					pchan1 <- *packet
				}
			case "send2":
				{
					packet := mcast.NewPacket(mcast.SND, 1, 10.0, 1)
					pchan2 <- *packet
				}
			default:
				{
				}
			}
		case packet1 := <-rchan1: // proc1からのメッセージを受信
			{
				fmt.Println("from process1")
				//				fmt.Printf("type: %d, sender: %d, rtime: %f, reqId: %d\n", packet.GetType(),
				//					packet.GetSender(), packet.GetTime(), packet.GetReqId())
				switch packet1.GetType() {
				case mcast.REQ:
					{
						fmt.Println("req from proc1 to mulcast in main")
						pchan1 <- packet1
						go func() {
							time.Sleep(delay * time.Second)
							pchan2 <- packet1
						}()
					}
				case mcast.ACK:
					{

						fmt.Printf("ack from proc1 in main: %1.1f\n", packet1.GetTime())
						pchan1 <- packet1
						// sleep and send pchan2
						go func() {
							time.Sleep(delay * time.Second)
							pchan2 <- packet1
						}()
					}
				default:
					{
						os.Exit(1)
					}
				}
			}
		case packet2 := <-rchan2: // proc2からのメッセージを受信
			{
				fmt.Println(sp + "from process2")
				switch packet2.GetType() {
				case mcast.REQ:
					{
						fmt.Println(sp + "req from proc2 to mulcast in main")
						pchan2 <- packet2
						go func() {
							time.Sleep(delay * time.Second)
							pchan1 <- packet2
						}()
					}
				case mcast.ACK:
					{
						fmt.Printf(sp+"ack from proc2 in main: %1.1f\n", packet2.GetTime())
						pchan2 <- packet2
						// sleep and send pchan2
						go func() {
							time.Sleep(delay * time.Second)
							pchan1 <- packet2
						}()
					}
				default:
					{
						os.Exit(1)
					}
				}
			}
		}
	}
}

func main() {
	mcan := make(chan string)
	go run(mcan)
	defer close(mcan)
	sc := bufio.NewScanner(os.Stdin)
	for {
		if sc.Scan() {
			fmt.Println(sc.Text())
			mcan <- sc.Text()
		}
	}
}
