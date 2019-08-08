package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/DMEvanCT/ProtocolBuffer/todo"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"os"
)





func main() {
   srv := grpc.NewServer()
   var tasks taskServer
   todo.RegisterTasksServer(srv, tasks)
   l, err := net.Listen("tcp", ":8888")
   if err != nil {
   	log.Fatal("Unable to listgen on 8888: %v", err)
   }
   log.Fatal(srv.Serve(l))
}

type taskServer struct {

}
type length int64

const (
	dbPath="mydb.pb"
	sizeOfLength = 8
)

var endianness = binary.LittleEndian



func (taskServer) Add( ctx context.Context, text *todo.Text) (*todo.Task, error) {
	task := &todo.Task{
		Text:                 text.Text,
		Done:                 false,

	}
	b, err := proto.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("Could not encode task: %v", err)
	}

	f, err := os.OpenFile(dbPath, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s: file: %v", dbPath, err)
	}
	if err := gob.NewEncoder(f).Encode(int64(len(b))); err != nil {
		return nil,  fmt.Errorf("Could not encode lenghth of prot: %v", err)

	}
	_, err = f.Write(b)
	if err != nil {
		return nil,  fmt.Errorf("could not write tastk to file: %v", err )
	}

	if err := f.Close();
	err != nil {
		return nil, fmt.Errorf("could not close file: %v", err)
	}

	return task, nil
}


func (taskServer) List(ctx context.Context, void *todo.Void) (*todo.TaskList, error) {
	b, err := ioutil.ReadFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", dbPath, err)
	}

	var tasks todo.TaskList
	for {
		log.Println(b)
		if len(b) == 0 {
			return &tasks, nil
		} else if len(b) < sizeOfLength {
			return nil, fmt.Errorf("remaining odd %d bytes, what to do?", len(b))
		}

		var l length
		if err := binary.Read(bytes.NewReader(b[:sizeOfLength]), endianness, &l); err != nil {
			return nil, fmt.Errorf("could not decode message length: %v", err)
		}
		//b = b[sizeOfLength:]
		log.Print(b)

		var task todo.Task
		if err := proto.Unmarshal(b[:l], &task); err != nil {
			return nil, fmt.Errorf("could not read task: %v", err)
		}
		b = b[l:]
		tasks.Tasks = append(tasks.Tasks, &task)
	}
}