package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	slotTimeInterval      = 2
	activeConnectionBase  = 0
	passiveConnectionBase = 1000
)

var (
	numberOfNodes     int64
	numberOfDelegates int64
)

func gobInterfaceRegister() {
	gob.Register(Block{})
	gob.Register(Transaction{})
}

func init() {
	gobInterfaceRegister()
}

func main() {
	//var nodeID string
	var nodeID int64
	var listenPort int64

	//flag.StringVar(&nodeID, "i", "", "Setup current node ID")
	flag.Int64Var(&nodeID, "i", 0, "Setup current node ID")
	flag.Int64Var(&listenPort, "p", 0, "Setup service port")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if listenPort == 0 {
		fmt.Println("Pick node ID and listening port")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigInt := make(chan os.Signal, 1)
	sysDone := make(chan struct{}, 1)
	signal.Notify(sigInt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigInt
		sysDone <- struct{}{}
	}()

	fmt.Println()

	node := NewNode(ctx, nodeID, listenPort)
	time.Sleep(time.Second * 3)

	node.Connect(ctx)
	time.Sleep(time.Second * 3)

	//i := 0
	//for key := range node.Peers {
	//	fmt.Println("Node[", node.ID, "]:", "[", i, "]", node.Peers[key])
	//	i++
	//}

	go node.StartForging()

	<-sysDone
	cancel()
}
