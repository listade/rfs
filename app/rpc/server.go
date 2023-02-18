package rpc

import (
	"context"
	"io"
	"os"
)

const (
	PageSize = 1024 * 1024 // 1 Mb
)

func NewServer(filename string) MemoryServerServer {
	return &server{
		filename,
	}
}

type server struct {
	filename string
}

func (s *server) Write(ctx context.Context, arg *WriteArg) (r *WriteRes, err error) {
	r = &WriteRes{}
	f, err := os.OpenFile(s.filename, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Seek(arg.GetPageNum()*PageSize, io.SeekStart)
	if err != nil {
		return
	}
	_, err = f.Write(arg.GetData())
	return
}

func (s *server) Read(ctx context.Context, arg *ReadArg) (r *ReadRes, err error) {
	r = &ReadRes{Bytes: make([]byte, PageSize)}
	f, err := os.Open(s.filename)
	if err != nil {
		return
	}
	_, err = f.Seek(arg.GetPageNum()*PageSize, io.SeekStart)
	if err != nil {
		return
	}
	_, err = f.Read(r.Bytes)
	return
}
