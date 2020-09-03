package main

import (
	"fmt"
	"github.com/trustmaster/goflow"
	"strconv"
)

type Ping struct {
	StartTrigger <-chan string

	In  <-chan string
	Out chan<- string
}

func (p *Ping) Process() {
	count := 0
	stop := false
	for {
		select {
		case msg, ok := <-p.StartTrigger:
			if ok {
				fmt.Println("Start Triggered Received", msg)
				p.Out <- msg
			}
		case msg, ok := <-p.In:
			if ok {
				//fmt.Println("Ping.In Received")
				if msg == "STOP" {
					fmt.Println("STOP Received")
					stop = true
				}
				count++
				p.Out <- msg + ".Pi-" + strconv.Itoa(count)
			}
		}
		if stop {
			break
		}
	}
	fmt.Println("Ping Ended")
}

type Pong struct {
	In  <-chan string
	Out chan<- string

	max int
}

func (p *Pong) Process() {
	count := 0
	for msg := range p.In {
		//fmt.Println("Pong.In Received")
		count++
		outStr := msg + ".Po-" + strconv.Itoa(count)
		p.Out <- outStr
		if count >= p.max {
			fmt.Println("STOPPING")
			p.Out <- "STOP"
			break
		} else {
			fmt.Println(outStr)
		}
	}
	fmt.Println("Pong Ended")
}

func NewPingPongApp() *goflow.Graph {
	n := goflow.NewGraph()

	n.Add("ping", &Ping{})
	n.Add("pong", &Pong{max: 10})
	n.ConnectBuf("ping", "Out", "pong", "In", 2)
	n.ConnectBuf("pong", "Out", "ping", "In", 2)
	n.MapInPort("GIn", "ping", "StartTrigger")

	return n
}

func main() {
	fmt.Println("Ping pong started")

	net := NewPingPongApp()
	in := make(chan string, 2)
	net.SetInPort("GIn", in)
	wait := goflow.Run(net)

	in <- "S"
	close(in)
	<-wait
}
