package main

import (
	"fmt"
	"net"
	"bufio"
	"os"
	"github.com/junabbott/messenger/pb"
	"strconv"
	"strings"
	"io"
)

const SRVR_ADDR string = ":8080"

func receiver(conn net.Conn) {
	for {
		response := &pb.Response{}
		err := pb.ReadAndUnmarshal(conn, response)
		if err == io.EOF {
			fmt.Println("Hub offline. Please try it again later.\nShutting down...")
			os.Exit(0)
		} else {
			fmt.Printf("Response: %+v\n", response)
		}
	}
}

func main() {
	fmt.Println("Connecting to " + SRVR_ADDR)
	conn, err := net.Dial("tcp", SRVR_ADDR)
	if err != nil {
		fmt.Printf("Error while connecting: %v\n", err)
		return
	}
	defer func(){
		fmt.Println("Shutting down...")
		conn.Close()
	}()
	fmt.Println("Connected")
	go receiver(conn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Options: 1) ID 2) List 3) Relay")

	for {
		option, _ := reader.ReadString('\n')
		switch strings.TrimRight(option, "\n") {
		case "1":
			request := &pb.Request{ RequestType: pb.Request_ID }
			out, _ := pb.Marshal(request)
			conn.Write(out)
		case "2":
			request := &pb.Request{ RequestType: pb.Request_LIST }
			out, _ := pb.Marshal(request)
			conn.Write(out)
		case "3":
			fmt.Println("Please enter user_ids (separated by comma): ")
			userIds, _ := reader.ReadString('\n')
			splits := strings.Split(strings.TrimRight(userIds, "\n"), ",")
			userIdsInts := make([]uint64, len(splits))
			for i, str := range splits {
				out, _ := strconv.ParseUint(strings.TrimSpace(str), 10, 64)
				userIdsInts[i] = out
			}
			fmt.Println("Message type (text, json, binary, unknown): ")
			dType,_ := reader.ReadString('\n')
			dataType :=  strings.TrimRight(dType, "\n")
			fmt.Println("Message: ")
			data,_ := reader.ReadString('\n')
			pb.RelayMessage(conn, userIdsInts, dataType, []byte(strings.TrimRight(data, "\n")))
		default:
			fmt.Println("Invalid option. Please try it again.")
		}
	}
}