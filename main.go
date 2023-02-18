package main

import (
	"flag"
	"fmt"
	"io"
	"labfs/app/fs"
	"labfs/app/rpc"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
)

var (
	l = flag.Bool("l", false, "list")
	w = flag.Bool("w", false, "write")
	r = flag.Bool("r", false, "read")
	d = flag.Bool("d", false, "delete")
	t = flag.Bool("t", false, "trunc")
	o = flag.Int("o", 0, "offset")

	s = flag.String("s", "", "server")
	h = flag.String("h", "127.0.0.1", "host")
	p = flag.Int("p", 5000, "port")
)

func main() {
	flag.Parse()

	if *s != "" {
		t, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *h, *p))
		if err != nil {
			log.Fatal(err)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		srv := grpc.NewServer()
		msrv := rpc.NewServer(*s)

		rpc.RegisterMemoryServerServer(srv, msrv)

		go func() {
			<-c
			srv.GracefulStop()
		}()
		err = srv.Serve(t)

		if err != nil {
			log.Fatal(err)
		}
		return
	}

	s, err := rpc.NewServerCloser()
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	fs := fs.NewRFS(s)

	if *l {
		l, err := fs.List()
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range l {
			fmt.Println(v)
		}
		return
	}
	if *d {
		err = fs.Delete(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	f, err := fs.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Seek(int64(*o), io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
	if *r {
		_, err = io.Copy(os.Stdout, f)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if *w {
		n, err := io.Copy(f, os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(n)
		return
	}
	if *t {
		f.Trunc()
	}
}
