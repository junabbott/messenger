package pb

import (
	"github.com/golang/protobuf/proto"
	"encoding/binary"
	"net"
	"fmt"
	"io"
	"strings"
)

type Message interface {
	Reset()
	String() string
	ProtoMessage()
	Descriptor() ([]byte, []int)
}

func Marshal(msg Message) ([]byte, error) {
	out, err := proto.Marshal(msg)
	if err != nil {
		fmt.Printf("error %v\n", err)
	}
	length := uint32(len(out))
	lengthInBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthInBytes, length)
	return append(lengthInBytes, out...), err
}

func RelayMessage(conn net.Conn, userIds []uint64, dType string, data []byte) {
	if len(data) > 1024000 {
		fmt.Println("Data must be less than 1024KB.")
		return
	}
	fmt.Printf("Relaying %v bytes of data\n", len(data))
	maxUserIds := len(userIds)
	if maxUserIds > 255 {
		fmt.Println("Maximum 255 user_ids has reached. The message will be sent to the first 255 users.")
		maxUserIds = 255
	}
	relayRequest := &RelayRequest{ userIds[:maxUserIds], MsgType(MsgType_value[strings.ToUpper(dType)]), data }
	request := &Request{ RequestType: Request_RELAY, RelayRequest: relayRequest}
	out, _ := Marshal(request)
	conn.Write(out)
}

func ReadAndUnmarshal(conn net.Conn, msg Message) error {
	lengthInBytes := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthInBytes)
	if err == io.EOF {
		conn.Close()
		return err
	}
	length := binary.LittleEndian.Uint32(lengthInBytes)
	fmt.Printf("Reading data %v bytes\n",length)
	data := make([]byte, length)
	_, errr := io.ReadFull(conn, data)
	if errr == io.EOF {
		conn.Close()
		return errr
	}
	proto.Unmarshal(data, msg)
	return nil
}