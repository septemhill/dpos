package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type ConnectionCodec struct {
	Enc *gob.Encoder
	Dec *gob.Decoder
}

type Node struct {
	ID       string
	Peers    map[string]*Peer
	Listener *net.TCPListener
	Chain    *Blockchain
	RWMutex  sync.RWMutex
	LastSlot int64
	Conns    map[*net.TCPConn]*ConnectionCodec
}

var peerInfos = make([]PeerInfo, 0)

func loadRemoteAddressList() {
	type peerinfos struct {
		Peers []PeerInfo
	}

	var infos peerinfos
	fd, err := os.Open("./config.json")
	if err != nil {
		fmt.Println("[Open Failed]", err)
	}

	byteArray, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Println("[ReadAll Failed]", err)
	}

	err = json.Unmarshal(byteArray, &infos)

	if err != nil {
		fmt.Println("[Unmarshal Failed]", err)
	}

	peerInfos = append(peerInfos, infos.Peers...)

	numberOfNodes = int64(len(peerInfos))
	numberOfDelegates = int64(numberOfNodes)
}

func init() {
	loadRemoteAddressList()
}

func handleConnection(ctx context.Context, conn *net.TCPConn, node *Node) {
	dec := node.Conns[conn].Dec
	for {
		var msg Message
		conn.SetReadDeadline(time.Now().Add(time.Second * 1))
		err := ReceiveMessage(&msg, dec)

		fmt.Println("[RECEIVED MSG 1]")
		if err != nil {
			fmt.Println("[RECEIVED MSG 2]", err)
			if err.Error() == "EOF" {
				fmt.Println(err)
				conn.Close()
			}
			//fmt.Println(err)
			//if err.(*net.OpError).Err.Error() != "i/o timeout" {
			//	log.Println("[Decode Failed]", err)
			//}
		} else {
			fmt.Println("[RECEIVED MSG 3]", msg)
			node.ProcessMessage(ctx, &msg, conn)
		}

		select {
		case <-ctx.Done():
			conn.Close()
		default:
			time.Sleep(time.Microsecond * 20)
		}
	}
}

func acceptConnection(ctx context.Context, listener *net.TCPListener, node *Node) {
	listenerCtx, cancel := context.WithCancel(context.Background())

	for {
		listener.SetDeadline(time.Now().Add(time.Second * 1))
		conn, err := listener.AcceptTCP()

		if err != nil {
			if err.(*net.OpError).Err.Error() != "i/o timeout" {
				log.Println("[AcceptTCP Failed]", err)
			}
		} else {
			codec := &ConnectionCodec{
				gob.NewEncoder(conn),
				gob.NewDecoder(conn),
			}

			node.Conns[conn] = codec
			go handleConnection(listenerCtx, conn, node)
		}

		select {
		case <-ctx.Done():
			cancel()
			listener.Close()
			return
		default:
			time.Sleep(time.Millisecond * 20)
		}
	}
}

func newServer(ctx context.Context, listenPort int64, node *Node) *net.TCPListener {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ":"+strconv.FormatInt(listenPort, 10))
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		log.Println("[ListenTCP Failed]", err)
		return nil
	}

	go acceptConnection(ctx, listener, node)

	return listener
}

//func NewNode(ctx context.Context, nodeID string, port int64) *Node {
func NewNode(ctx context.Context, nodeID int64, port int64) *Node {
	node := &Node{
		//ID:    nodeID,
		ID:    peerInfos[nodeID].ID,
		Peers: make(map[string]*Peer, 0),
		Conns: make(map[*net.TCPConn]*ConnectionCodec, 0),
	}

	node.Listener = newServer(ctx, port, node)
	node.Chain = NewBlockchain(node)

	return node
}

func (n *Node) Connect(ctx context.Context) {
	for i := int64(0); i < numberOfDelegates; i++ {
		if peerInfos[i].ID != n.ID {
			n.RWMutex.Lock()
			_, ok := n.Peers[peerInfos[i].ID]

			if !ok {
				tcpAddr, _ := net.ResolveTCPAddr("tcp", peerInfos[i].Destination)
				conn, err := net.DialTCP("tcp", nil, tcpAddr)

				if err != nil {
					fmt.Println("[DialTCP Failed]", err)
					n.RWMutex.Unlock()
					continue
				}

				codec := &ConnectionCodec{
					gob.NewEncoder(conn),
					gob.NewDecoder(conn),
				}

				n.Conns[conn] = codec

				peer := NewPeer(ctx, n, conn, peerInfos[i].ID, true)
				n.Peers[peerInfos[i].ID] = peer
			}
			n.RWMutex.Unlock()
		}
	}
}

func (n *Node) StartForging() {
	for {
		currentSlot := GetSlotNumber(0)
		lastBlock := n.Chain.GetLastBlock()
		lastSlot := GetSlotNumber(GetTime(lastBlock.GetTimestamp()))

		if currentSlot == lastSlot {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if currentSlot == n.LastSlot {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		delegateSlot := currentSlot % numberOfDelegates
		delegateID := peerInfos[delegateSlot].ID

		if delegateID == n.ID {
			newBlock := n.Chain.CreateBlock()
			n.Broadcast(BlockMessage( /*n.ID, */ *newBlock))
			fmt.Println("[NODE", n.ID, " NewBlock]", newBlock)
			n.Chain.CommitBlock(newBlock)
			fmt.Println("====================[END OF NEW BLOCK]====================")
			n.LastSlot = currentSlot
		}

		time.Sleep(time.Second * 1)
	}
}

func (n *Node) Broadcast(msg *Message) {
	for _, peer := range n.Peers {
		SendMessage(msg, n.Conns[peer.Conn].Enc, n.ID)
	}
}

func (n *Node) handleInitMessage(ctx context.Context, msg *Message, conn *net.TCPConn) {
	nodeID := msg.Body.(string)

	n.RWMutex.Lock()
	_, ok := n.Peers[nodeID]
	if !ok {
		peer := NewPeer(ctx, n, conn, nodeID, false)
		n.Peers[nodeID] = peer
	} else {
		delete(n.Conns, conn)
		conn.Close()
	}
	n.RWMutex.Unlock()
}

func (n *Node) handleBlockMessage(msg *Message) {
	block := msg.Body.(Block)
	fmt.Println("[Node", n.ID, "Receive Block]", block)
	n.Chain.CommitBlock(&block)
	fmt.Println("====================[END OF RECEIVE BLOCK]====================")
}

func (n *Node) ProcessMessage(ctx context.Context, msg *Message, conn *net.TCPConn) {
	switch msg.Type {
	case MessageTypeInit:
		n.handleInitMessage(ctx, msg, conn)
	case MessageTypeBlock:
		n.handleBlockMessage(msg)
	default:
		fmt.Println("[ProcessMessage Failed] Unknown Message Type")
	}
}
