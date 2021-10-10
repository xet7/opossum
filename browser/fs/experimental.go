package fs

import (
	"bytes"
	"fmt"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
	"github.com/psilva261/opossum/logger"
	"github.com/psilva261/opossum/nodes"
	"golang.org/x/net/html"
)

// Node such that one obtains a file structure like
//
// /0
// /0/attrs
// /0/html
// /0/tag
// /0/0
//   ...
// /0/1
//   ...
// ...
//
// (dir structure stolen from domfs)
type Node struct {
	name string
	nt *nodes.Node
}

func (n Node) Stat() (s proto.Stat) {
	s = *oFS.NewStat(n.name, un, gn, 0700)
	s.Mode |= proto.DMDIR
	// qtype bits should be consistent with Stat mode.
	s.Qid.Qtype = uint8(s.Mode >> 24)
	return
}

func (n Node) WriteStat(s *proto.Stat) error {
	return nil
}

func (n Node) SetParent(p fs.Dir) {
}

func (n Node) Parent() fs.Dir {
	return nil
}

func (n Node) Children() (cs map[string]fs.FSNode) {
	cs = make(map[string]fs.FSNode)
	if n.nt == nil {
		return
	}
	for i, c := range n.nt.Children {
		ddn := fmt.Sprintf("%v", i)
		cs[ddn] = &Node{
			name: ddn,
			nt: c,
		}
	}
	if n.nt.Type() == html.ElementNode {
		cs["tag"] = n.tag()
		cs["attrs"] = Attrs{attrs: &n.nt.DomSubtree.Attr}
		cs["html"] = n.html()
	}

	return
}

func (n Node) tag() fs.FSNode {
	return fs.NewDynamicFile(
		oFS.NewStat("tag", un, gn, 0666),
		func() []byte {
			return []byte(n.nt.Data())
		},
	)
}

func (n Node) html() fs.FSNode {
	return fs.NewDynamicFile(
		oFS.NewStat("html", un, gn, 0666),
		func() []byte {
			buf := bytes.NewBufferString("")
			if err := html.Render(buf, n.nt.DomSubtree); err != nil {
				log.Errorf("render: %v", err)
				return []byte{}
			}
			return []byte(buf.String())
		},
	)
}

func (n Node) AddChild(fs.FSNode) error {
	return nil
}

func (n Node) DeleteChild(name string) error {
	return fmt.Errorf("no removal possible")
}

type Attrs struct {
	attrs *[]html.Attribute
}

func (as Attrs) Stat() (s proto.Stat) {
	s = *oFS.NewStat("attrs", un, gn, 0500)
	s.Mode |= proto.DMDIR
	// qtype bits should be consistent with Stat mode.
	s.Qid.Qtype = uint8(s.Mode >> 24)
	return
}

func (as Attrs) WriteStat(s *proto.Stat) error {
	return nil
}

func (as Attrs) SetParent(p fs.Dir) {
}

func (as Attrs) Parent() fs.Dir {
	return nil
}

func (as Attrs) Children() (cs map[string]fs.FSNode) {
	log.Infof("Attrs#Children()")
	cs = make(map[string]fs.FSNode)
	ff := func(k string) fs.FSNode {
		return fs.NewDynamicFile(
			oFS.NewStat(k, un, gn, 0666),
			func() []byte {
				var v string
				for _, a := range *as.attrs {
					if a.Key == k {
						v = a.Val
					}
				}
				return []byte(v)
			},
		)
	}
	for _, attr := range *as.attrs {
		cs[attr.Key] = ff(attr.Key)
	}
	return 
}
