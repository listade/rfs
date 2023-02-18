package fs

import (
	"context"
	"io"
	"labfs/app/rpc"
)

type File interface {
	io.ReadWriteSeeker
	io.Closer
	Trunc()
}

type file struct {
	fa  *attr
	fs  *fs
	pos int64
}

func (f *file) getPage() (p int64, err error) {
	if f.fa.Page == noPage {
		p, err = f.fs.ft.Pages.malloc()
		f.fa.Page = p
		return
	}
	p = f.fa.Page
	for i := 0; i < int(f.pos)/rpc.PageSize; i++ {
		n := f.fs.ft.Pages[p]
		if n == noPage {
			n, err = f.fs.ft.Pages.malloc()
			if err != nil {
				return
			}
			f.fs.ft.Pages[p] = n
		}
		p = n
	}
	return
}

func (f *file) Read(b []byte) (n int, err error) {
	for {
		if f.pos >= f.fa.Size {
			n -= int(f.pos) - int(f.fa.Size)
			err = io.EOF
			return
		}
		if n == len(b) {
			return
		}
		var p int64
		p, err = f.getPage()
		if err != nil {
			return
		}
		var res *rpc.ReadRes
		res, err = f.fs.mc.Read(context.Background(), &rpc.ReadArg{PageNum: p})
		if err != nil {
			return
		}
		rb := res.GetBytes()
		off := f.pos % rpc.PageSize
		cb := copy(b[n:], rb[off:])
		n += cb
		f.pos += int64(cb)
	}
}

func (f *file) Write(b []byte) (n int, err error) {
	for {
		if n == len(b) {
			return
		}
		var p int64
		p, err = f.getPage()
		if err != nil {
			return
		}
		var res *rpc.ReadRes
		res, err = f.fs.mc.Read(context.Background(), &rpc.ReadArg{PageNum: p})
		if err != nil {
			return
		}
		rb := res.GetBytes()
		off := f.pos % rpc.PageSize
		cb := copy(rb[off:], b[n:])

		_, err = f.fs.mc.Write(context.Background(), &rpc.WriteArg{PageNum: p, Data: rb[:]})
		if err != nil {
			return
		}
		n += cb
		f.pos += int64(cb)
		if f.pos >= f.fa.Size {
			f.fa.Size = f.pos
		}
	}
}

const (
	SEEK_SET = iota
	SEEK_CUR
	SEEK_END
)

func (f *file) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case SEEK_SET:
		f.pos = offset
	case SEEK_CUR:
		f.pos += offset
	case SEEK_END:
		f.pos = f.fa.Size + offset - 1
	}
	n = f.pos
	return
}

func (f *file) currentPage() (p int64) {
	p = f.fa.Page
	for i := 0; i < int(f.pos)/rpc.PageSize; i++ {
		var ok bool
		p, ok = f.fs.ft.Pages[p]
		if !ok || p == noPage {
			return
		}
	}
	return
}

func (f *file) Trunc() {
	if f.pos < f.fa.Size {
		f.fa.Size = f.pos
	}
	p := f.currentPage()
	if p == noPage {
		return
	}
	var ok bool
	p, ok = f.fs.ft.Pages[p]
	if !ok {
		return
	}
	for {
		if p == noPage {
			return
		}
		n := p
		p, ok = f.fs.ft.Pages[p]
		delete(f.fs.ft.Pages, n)
		if !ok {
			return
		}
	}
}

func (f *file) Close() error {
	return f.fs.ft.save(fsTableName)
}
