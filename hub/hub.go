package main

import (
	"fmt"
	"net"
	"sync"
	"math/rand"
	"strconv"
	"strings"
	"github.com/junabbott/messenger/pb"
	"io"
)

const SRVR_ADDR string = ":8080"

type Clients struct {
	sync.RWMutex
	m map[uint64]net.Conn
}

var clients = Clients{m: make(map[uint64]net.Conn)}

func (* Clients) remove(userId uint64) {
	clients.Lock()
	delete(clients.m, userId)
	clients.Unlock()
}

func (* Clients) add(userId uint64, conn net.Conn) {
	clients.Lock()
	clients.m[userId] = conn
	clients.Unlock()
}

func (* Clients) contains(userId uint64) bool {
	clients.RLock()
	_, ok := clients.m[userId]
	clients.RUnlock()
	return ok
}

func getRandomUserId() uint64 {
	user_id := rand.Uint64()
	alreadyExists := clients.contains(user_id)
	for alreadyExists {
		user_id = rand.Uint64()
		alreadyExists = clients.contains(user_id)
	}
	return user_id
}

func handleConnection(conn net.Conn) {
	user_id := getRandomUserId()
	fmt.Printf("User %v has been connected\n", user_id)
	clients.add(user_id, conn)
	for {
		request := &pb.Request{}
		err := pb.ReadAndUnmarshal(conn, request)
		if err == io.EOF {
			clients.remove(user_id)
			fmt.Printf("User %v has been disconnected\n", user_id)
			return
		}
		switch request.RequestType {
		case pb.Request_ID:
			response := &pb.Response { MsgType: pb.MsgType_TEXT, Msg: []byte(strconv.FormatUint(user_id, 10)) }
			out, _ := pb.Marshal(response)
			fmt.Printf("Sending ID to %v\n", user_id)
			conn.Write(out)
		case pb.Request_LIST:
			newMessage := ""
			clients.RLock()
			for k, _ := range clients.m {
				if k != user_id {
					newMessage += strconv.FormatUint(k, 10) + ","
				}
			}
			clients.RUnlock()
			response := &pb.Response { MsgType: pb.MsgType_TEXT, Msg: []byte(strings.TrimRight(newMessage,",")) }
			out, _ := pb.Marshal(response)
			fmt.Printf("Sending LIST to %v\n", user_id)
			conn.Write(out)
		case pb.Request_RELAY:
			relayMessage(user_id, *request.RelayRequest)
		default:
			response := &pb.Response { Msg: []byte("Got your message") }
			out, _ := pb.Marshal(response)
			conn.Write(out)
		}
	}
}

func relayMessage(from uint64, msg pb.RelayRequest) {
	for _, id := range msg.UserIds {
		clients.RLock()
		if conn, ok := clients.m[id]; ok {
			response := &pb.Response {from, msg.MsgType, msg.Msg }
			out, _ := pb.Marshal(response)
			_, err := conn.Write(out)
			if err == io.EOF {
				clients.remove(id)
				fmt.Printf("User %v has been disconnected\n", id)
			}
		} else {
			fmt.Printf("Skipping user %v that doesn't exist\n", id)
		}
		clients.RUnlock()
		fmt.Printf("Sending RELAY to %v\n", id)
	}
}

func main() {
	fmt.Println("Init Hub " + SRVR_ADDR)
	ln, err := net.Listen("tcp", SRVR_ADDR)
	if err != nil {
		fmt.Printf("Error while init server: %v\n", err)
	}
	defer func(){
		fmt.Println("Shutting down...")
		ln.Close()
	}()
	fmt.Println("Hub Online")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error while accepting conn: %v\n", err)
		}
		go handleConnection(conn)
	}
}