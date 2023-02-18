package rpc

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServerCloser interface {
	MemoryServerClient
	io.Closer
}

func NewServerCloser() (s ServerCloser, err error) {
	var hosts []string
	var b []byte
	b, err = ioutil.ReadFile("hosts.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &hosts)
	if err != nil {
		return
	}
	s, err = NewServersMediator(hosts...)
	if err != nil {
		return
	}
	return
}

func NewServersMediator(hosts ...string) (m ServerCloser, err error) {
	cs := make([]*grpc.ClientConn, len(hosts))
	m = mediator(cs)
	for k, v := range hosts {
		var c *grpc.ClientConn
		c, err = grpc.Dial(v, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		cs[k] = c
	}
	return
}

type mediator []*grpc.ClientConn

func (m mediator) recalc(n int64) (c *grpc.ClientConn, p int64) {
	hosts := int64(len(m))
	c = m[n%hosts]
	p = n / hosts
	return
}

func (m mediator) Write(ctx context.Context, in *WriteArg, opts ...grpc.CallOption) (*WriteRes, error) {
	c, p := m.recalc(in.GetPageNum())
	in.PageNum = p

	return NewMemoryServerClient(c).Write(ctx, in, opts...)
}

func (m mediator) Read(ctx context.Context, in *ReadArg, opts ...grpc.CallOption) (*ReadRes, error) {
	c, p := m.recalc(in.GetPageNum())
	in.PageNum = p

	return NewMemoryServerClient(c).Read(ctx, in, opts...)
}

func (s mediator) Close() (err error) {
	for _, v := range s {
		err = v.Close()
	}
	return
}
