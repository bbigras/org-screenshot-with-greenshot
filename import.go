package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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

	log.Println("copy", t.OrgFilePath, path)
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

	f, err := os.OpenFile("d:\\testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	flag.Parse()

	log.SetOutput(f)

	fmt.Println("bleh")
	fmt.Println(len(os.Args), os.Args)

	switch {
	case *send != "":
		fmt.Println("client")
		fmt.Println(os.Args[1])

		client, err := rpc.DialHTTP("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
		if err != nil {
			log.Fatal("dialing:", err)
		}

		err = client.Call("Path.Send", *send, nil)
		if err != nil {
			log.Fatal("arith error:", err)
		}
	default:
		fmt.Println("server")

		path := &Path{
			OrgFilePath: os.Args[1],
			ChanQuit:    chanQuit,
		}

		rpc.Register(path)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if e != nil {
			log.Fatal("listen error:", e)
		}
		go http.Serve(l, nil)
		<-chanQuit
	}
}
