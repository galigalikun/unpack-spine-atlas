package unpackspineatlas

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Parser struct {
	path string
	r    *bufio.Reader
	file *os.File
}

func (p *Parser) NewAtlas() (atlas *Atlas, err error) {
	atlas = &Atlas{
		Images: make([]*Image, 0),
	}
	offset := int64(0)
	for {
		b, _, err := p.r.ReadLine()
		if err == io.EOF {
			break
		}
		if len(b) == 0 {
			continue
		}
		line := string(b)
		a := strings.Split(line, ":")
		if len(a) == 1 {
			if len(atlas.Image) == 0 {
				atlas.Image = filepath.Join(p.path, line)
			} else {
				p.file.Seek(offset, os.SEEK_SET)
				p.r.Reset(p.file)
				break
			}
		} else {
			switch a[0] {
			case "size":
				atlas.Size = line
			case "format":
				atlas.Format = line
			case "filter":
				atlas.Filter = line
			case "repeat":
				atlas.Repeat = line
			}
		}
		offset += int64(len(b)) + 1
	}

	return
}

func (p *Parser) Parse() ([]*Atlas, error) {
	var atlas []*Atlas
	a, err := p.NewAtlas()
	if err != nil {
		return nil, err
	}
	for {
		b, _, err := p.r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line := string(b)

		log.Printf(line)

		if len(line) == 1 {
			atlas = append(atlas, a)
			a, err = p.NewAtlas()
			if err != nil {
				return nil, err
			}
			continue
		}

		a.NewImage(p.r, line)

	}
	atlas = append(atlas, a)

	return atlas, nil
}

func NewParser(s string) (*Parser, error) {
	file, err := os.Open(s)
	if err != nil {
		return nil, err
	}

	p := &Parser{
		path: filepath.Dir(s),
		file: file,
		r:    bufio.NewReader(file),
	}

	return p, nil
}

func (p *Parser) Close() {
	p.file.Close()
}
