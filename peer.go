package main

import (
	"context"
	"encoding/gob"
	"net"
)

type Peer struct {
	ID          string
	NodeID      string
	Conn        *net.TCPConn
	ConnEncoder *gob.Encoder
}

type PeerInfo struct {
	ID          string `json:"id"`
	Destination string `json:"dest"`
}

//func NewPassivePeer(ctx context.Context, node *Node) *Peer {
//	peer := &Peer{}
//}
//
//func NewActivePeer(ctx context.Context, node *Node, peerInfo PeerInfo) *Peer {
//	remoteAddr, _ := net.ResolveTCPAddr("tcp", peerInfo.Destination)
//	conn, err := net.DialTCP("tcp", nil, remoteAddr)
//
//	if err != nil {
//		log.Println("[DialTCP Failed]", err)
//		return nil
//	}
//
//	peer := &Peer{
//		ID:          peerInfo.ID,
//		NodeID:      node.ID,
//		Conn:        conn,
//		ConnEncoder: gob.NewEncoder(conn),
//	}
//
//	go handleConnection(ctx, conn, gob.NewDecoder(conn), node)
//
//	SendMessage(InitMessage(node.ID), peer.ConnEncoder, node.ID)
//
//	return peer
//}

func NewPeer(ctx context.Context, node *Node, conn *net.TCPConn, peerID string) *Peer {
	peer := &Peer{
		ID:          peerID,
		NodeID:      node.ID,
		Conn:        conn,
		ConnEncoder: gob.NewEncoder(conn),
	}

	go handleConnection(ctx, conn, gob.NewDecoder(conn), node)

	SendMessage(InitMessage(node.ID), peer.ConnEncoder, node.ID)

	return peer
}

//func NewPeer(ctx context.Context, node *Node, peerInfo PeerInfo) *Peer {
//	remoteAddr, _ := net.ResolveTCPAddr("tcp", peerInfo.Destination)
//	conn, err := net.DialTCP("tcp", nil, remoteAddr)
//
//	if err != nil {
//		log.Println("[DialTCP Failed]", err)
//		return nil
//	}
//
//	peer := &Peer{
//		ID:          peerInfo.ID,
//		NodeID:      node.ID,
//		Conn:        conn,
//		ConnEncoder: gob.NewEncoder(conn),
//	}
//
//	go handleConnection(ctx, conn, gob.NewDecoder(conn), node)
//
//	SendMessage(InitMessage(node.ID), peer.ConnEncoder, node.ID)
//
//	return peer
//}
