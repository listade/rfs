package fs

import (
	"encoding/json"
	"labfs/app/rpc"
	"os"
)

type FS interface {
	List() ([]string, error)
	Open(string) (File, error)
	Delete(string) error
}

type fs struct {
	mc rpc.MemoryServerClient
	ft *fsTable
}
type attr struct {
	Page int64
	Size int64
}
type fat map[string]*attr
type pat map[int64]int64

func (t pat) malloc() (p int64, err error) {
	for {
		_, ok := t[p]
		if !ok {
			t[p] = noPage
			return
		}
		p++
	}
}

type fsTable struct {
	Files fat
	Pages pat
}

const (
	fsTableName = "fstable.json"
	noPage      = -1
)

func (t *fsTable) save(file string) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return
	}
	e := json.NewEncoder(f)
	e.SetIndent("", " ")
	return e.Encode(t)
}

func (t *fsTable) restore(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	return json.NewDecoder(f).Decode(t)
}

func NewRFS(mc rpc.MemoryServerClient) FS {
	return &fs{
		mc,
		&fsTable{
			fat{},
			pat{},
		},
	}
}

func (s *fs) List() (l []string, err error) {
	s.ft.restore(fsTableName)
	defer s.ft.save(fsTableName)

	l = make([]string, len(s.ft.Files))
	i := 0
	for k := range s.ft.Files {
		l[i] = k
		i++
	}
	return
}

func (s *fs) Open(name string) (f File, err error) {
	s.ft.restore(fsTableName)
	fa, ok := s.ft.Files[name]
	if !ok {
		fp := &file{&attr{noPage, 0}, s, 0}
		s.ft.Files[name] = fp.fa
		f = fp
		return
	}
	f = &file{fa, s, 0}
	return
}

func (s *fs) Delete(name string) (err error) {
	s.ft.restore(fsTableName)
	defer s.ft.save(fsTableName)

	fa, ok := s.ft.Files[name]
	if !ok {
		err = os.ErrNotExist
		return
	}
	File(&file{fa, s, 0}).Trunc()

	delete(s.ft.Pages, fa.Page)
	delete(s.ft.Files, name)

	return
}
