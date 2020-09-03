package main

import (
	"fmt"
	"github.com/trustmaster/goflow"
	"strconv"
	"sync"
)

//  GraphIn -> ArrayPorts.In -> ArrayPorts.Out[0..2] (3x Fan-out)->(3x Fan-in) MapPorts.In[K0..K3] ->  MapPorts.Out -> Print

type ArrayPorts struct {
	In <-chan string
	Out []chan<- string
}

func (ap *ArrayPorts) Process() {
	wg := new(sync.WaitGroup)
	for s := range ap.In {
		for i := 0; i < len(ap.Out); i ++ {
			wg.Add(1)
			go func(idx int) {
				ap.Out[idx] <- s + "from GraphIn -> AP.In -> (fan-out) AP.Out[" + strconv.Itoa(idx) + "]"
				close(ap.Out[idx])
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
}


type MapPorts struct {
	In map[string]<-chan string
	Out chan<- string
}

func (mp *MapPorts) Process() {
	wg := new(sync.WaitGroup)
	for k, ch := range mp.In {
		k := k
		ch := ch
		wg.Add(1)
		go func() {
			for s := range ch {
				mp.Out <- s + " -> (fan-in) MP.In[" + k + "]"
			}
			wg.Done()
		}()
	}
	wg.Wait()
	//close(mp.Out)
}


type Print struct {
	Line <-chan string
}

func (p *Print) Process() {
	for s := range p.Line {
		fmt.Println(s + " -> Print.Line")
	}
}

func NewArrayMapApp() *goflow.Graph {
	n := goflow.NewGraph()
	n.Add("arrayports", new(ArrayPorts))
	n.Add("mapports", new(MapPorts))
	n.Add("print", new(Print))

	//  GraphIn -> ArrayPorts.In
	//
	n.MapInPort("GraphIn", "arrayports", "In")

	//  ArrayPorts.Out[0..2] (3x Fan-out)->(3x Fan-in) MapPorts.In[K0..K3]
	//
	n.Connect("arrayports", "Out[0]", "mapports", "In[K0]")
	n.Connect("arrayports", "Out[1]", "mapports", "In[K1]")
	n.Connect("arrayports", "Out[2]", "mapports", "In[K2]")

	//  MapPorts.Out -> Print
	//
	n.Connect("mapports", "Out", "print", "Line")

	return n
}

func main() {
	fmt.Println("ArrayMap started")
	net := NewArrayMapApp()
	in := make(chan string)
	net.SetInPort("GraphIn", in)
	wait := goflow.Run(net)

	in <- "S"
	close(in)
	<-wait
}