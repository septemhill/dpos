package main

import (
	"context"
	"net"
)

type Peer struct {
	ID     string
	NodeID string
	Conn   *net.TCPConn
	//ConnEncoder *gob.Encoder
}

type PeerInfo struct {
	ID          string `json:"id"`
	Destination string `json:"dest"`
}

func NewPeer(ctx context.Context, node *Node, conn *net.TCPConn, peerID string, initMsg bool) *Peer {
	peer := &Peer{
		ID:     peerID,
		NodeID: node.ID,
		Conn:   conn,
		//ConnEncoder: gob.NewEncoder(conn),
	}

	go handleConnection(ctx, conn /*gob.NewDecoder(conn),*/, node)

	//SendMessage(InitMessage(node.ID), peer.ConnEncoder, node.ID)
	if initMsg {
		SendMessage(InitMessage(node.ID), node.Conns[conn].Enc, node.ID)
	}

	return peer
}
