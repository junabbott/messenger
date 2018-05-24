package pb

import (
	"testing"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"reflect"
	"net"
	"math/rand"
	"io"
	"sync"
)

func TestMarshal(t *testing.T) {
	message := &Request{ RequestType: Request_ID }
	out, err := Marshal(message)
	if err != nil {
		t.Errorf("Error while marshalling the data: %v\n", err)
	}
	if len(out) == 0 {
		t.Error("The marshalled data should not be zero!")
	}
	lengthInBytes := out[0:4]
	length := binary.LittleEndian.Uint32(lengthInBytes)
	data := out[4:]
	if int(length) != len(data) {
		t.Errorf("The actual data size %v doesn't match expected data size %v\n", len(data), int(length))
	}
	actualMessage := &Request{}
	proto.Unmarshal(data, actualMessage)
	if !reflect.DeepEqual(message, actualMessage) {
		t.Error("The actual data doesn't match the expected data!")
	}
}

func TestRelayMessage(t *testing.T) {
	server, client := net.Pipe()
	userId := rand.Uint64()
	userIds := make([]uint64, 255)
	for i, _ := range userIds {
		userIds[i] = rand.Uint64()
	}
	wg := sync.WaitGroup{}
	largeData := make([]byte, 1024001)
	rand.Read(largeData)

	data := &Request{ RequestType: Request_RELAY, RelayRequest: &RelayRequest{userIds, MsgType_TEXT, largeData[:1024000]}}
	go func() {
		wg.Add(1)
		lengthInBytes := make([]byte, 4)
		_, err := io.ReadFull(server, lengthInBytes)
		if err != nil {
			t.Errorf("Error while reading first 4 bytes of the data: %v\n", err)
		}
		length := binary.LittleEndian.Uint32(lengthInBytes)
		actualData := make([]byte, length)
		_, errr := io.ReadFull(server, actualData)
		if errr != nil {
			t.Errorf("Error while reading the %v bytes of data: %v\n", length, errr)
		}
		request := &Request{}
		proto.Unmarshal(actualData, request)

		if !reflect.DeepEqual(request, data) {
			t.Error("The actual data does not match the expected data!")
		}
		server.Close()
		wg.Done()
	}()
	RelayMessage(client, data.RelayRequest.UserIds, "TEXT", largeData)
	RelayMessage(client, append(userIds, userId), "TEXT", largeData[:1024000])
	wg.Wait()
	client.Close()
}

func TestReadAndUnmarshal(t *testing.T) {
	server, client := net.Pipe()
	userId := rand.Uint64()
	request := &Request{ RequestType: Request_RELAY, RelayRequest: &RelayRequest{[]uint64{userId}, MsgType_TEXT, []byte("hello")}}
	data := &Request{}
	wg := sync.WaitGroup{}

	go func() {
		wg.Add(1)
		ReadAndUnmarshal(server, data)
		wg.Done()
	}()

	RelayMessage(client, request.RelayRequest.UserIds, "TEXT", request.RelayRequest.Msg)
	wg.Wait()
	if !reflect.DeepEqual(data, request) {
		t.Error("The actual data does not match the expected data!")
	}
}