package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

var (
	send = flag.String("send", "", "path of the capture to send")
	port = flag.Int("port", 64081, "TCP port to use")
)

type Path struct {
	OrgFilePath string
	ChanQuit    chan int
}

func (t *Path) Send(path string, reply *int) error {
	defer close(t.ChanQuit)

	_, errCopyFile := CopyFile(path, t.OrgFilePath)
	if errCopyFile != nil {
		return errCopyFile
	}

	return os.Remove(path)
}

func CopyFile(src, dst string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

func main() {
	chanQuit := make(chan int)

	flag.Parse()

	switch {
	case *send != "":
		client, err := rpc.DialHTTP("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
		if err != nil {
			panic(err)
		}

		err = client.Call("Path.Send", *send, nil)
		if err != nil {
			panic(err)
		}
	default:
		path := &Path{
			OrgFilePath: os.Args[1],
			ChanQuit:    chanQuit,
		}

		rpc.Register(path)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if e != nil {
			panic(e)
		}
		go http.Serve(l, nil)
		<-chanQuit
	}
}
