package main

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Parser struct {
	path string
	r    *bufio.Reader
	file *os.File
}

type Atlas struct {
	Image  string
	Size   string
	Format string
	Filter string
	Repeat string
	Images []Image
}

type Image struct {
	Name   string
	rotate bool
	xy     *image.Point
	size   *image.Point
	orig   *image.Point
	offset *image.Point
	index  int
}

func (i *Image) Rect() image.Rectangle {
	/*
	  xy: 1116, 1688
	  size: 846, 358
	*/
	return image.Rect(1116, 1688, 846, 358)
	if i.rotate {
		return image.Rect(i.xy.X, i.xy.Y, i.size.Y, i.size.X)
	}
	return image.Rect(i.xy.X, i.xy.Y, i.size.X, i.size.Y)
}

func (i *Image) parse(s string, err error, ch string) (string, error) {
	if err != nil {
		return "", err
	}
	ss := strings.Split(s, ":")
	if len(ss) != 2 {
		return "", fmt.Errorf("%s error", ch)
	}

	if strings.Index(ss[0], ch) == -1 {
		return "", fmt.Errorf("%s error", ch)
	}

	return ss[1], nil
}

func (i *Image) parseRotate(s string, err error) error {

	ss, e := i.parse(s, err, "rotate")
	if e != nil {
		return e
	}
	i.rotate = false
	if ss == "true" {
		i.rotate = true
	}
	return nil
}

func (i *Image) parsePoint(s string, err error, ch string) (*image.Point, error) {
	ss, e := i.parse(s, err, ch)
	if e != nil {
		return nil, e
	}
	point := strings.Split(ss, ",")
	if len(point) != 2 {
		return nil, fmt.Errorf("point %s error %s", ch, ss)
	}
	x, e := strconv.Atoi(strings.TrimSpace(point[0]))
	if e != nil {
		return nil, e
	}
	y, e := strconv.Atoi(strings.TrimSpace(point[0]))
	if e != nil {
		return nil, e
	}
	return &image.Point{X: x, Y: y}, nil
}

func (i *Image) parseXY(s string, err error) (e error) {

	i.xy, e = i.parsePoint(s, err, "xy")
	if e != nil {
		return e
	}
	return nil
}
func (i *Image) parseSize(s string, err error) (e error) {

	i.size, e = i.parsePoint(s, err, "size")
	if e != nil {
		return e
	}
	return nil
}
func (i *Image) parseOrig(s string, err error) (e error) {

	i.orig, e = i.parsePoint(s, err, "orig")
	if e != nil {
		return e
	}
	return nil
}

func (i *Image) parseOffset(s string, err error) (e error) {

	i.offset, e = i.parsePoint(s, err, "offset")
	if e != nil {
		return e
	}
	return nil
}
func (i *Image) parseIndex(s string, err error) (e error) {

	ss, e := i.parse(s, err, "index")
	if e != nil {
		return e
	}
	i.index, e = strconv.Atoi(strings.TrimSpace(strings.Trim(ss, "\n")))
	if e != nil {
		return e
	}
	return nil
}
func (a *Atlas) Unpack() error {
	file, err := os.Open(a.Image)
	if err != nil {
		return err
	}
	defer file.Close()

	atlasImage, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	createImage := make(chan error, len(a.Images))
	for _, img := range a.Images {
		go func() {
			newImg := atlasImage.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(img.Rect())
			os.MkdirAll(filepath.Dir(img.Name), 0755)
			out, err := os.Create(img.Name)
			if err != nil {
				createImage <- fmt.Errorf("create dir error")
				return
			}
			err = png.Encode(out, newImg)
			if err != nil {
				createImage <- fmt.Errorf("crate png error")
				return
			}
			createImage <- nil
		}()
	}

	for _ = range a.Images {
		res, ok := <-createImage

		fmt.Printf("end:%v:%v\n", res, ok)
	}

	return nil
}
func (a *Atlas) NewImage(r *bufio.Reader, line string) (err error) {
	image := Image{
		Name: fmt.Sprintf("%s.png", strings.Trim(line, "\n")),
	}
	err = image.parseRotate(r.ReadString('\n'))
	if err != nil {
		fmt.Printf("e2:%v:%s\n", err, line)
		return err
	}
	err = image.parseXY(r.ReadString('\n'))
	if err != nil {
		fmt.Printf("e3:%v\n", err)
		return err
	}
	err = image.parseSize(r.ReadString('\n'))
	if err != nil {
		fmt.Printf("e4:%v", err)
		return err
	}
	err = image.parseOrig(r.ReadString('\n'))
	if err != nil {
		fmt.Printf("e5:%s:%v", err, line)
		return err
	}
	err = image.parseOffset(r.ReadString('\n'))
	if err != nil {
		fmt.Printf("e6:%v", err)
		return err
	}
	err = image.parseIndex(r.ReadString('\n'))
	if err != nil {
		fmt.Printf("e7:%v", err)
		return err
	}

	a.Images = append(a.Images, image)

	return nil
}

func (p *Parser) NewAtlas() (atlas *Atlas, err error) {
	atlas = &Atlas{}

	img, err := p.r.ReadString('\n')
	atlas.Image = fmt.Sprintf("%s/%s", p.path, strings.Trim(img, "\n"))
	if err != nil {
		fmt.Printf("a1:%v", err)
		return
	}
	atlas.Size, err = p.r.ReadString('\n')
	if err != nil {
		fmt.Printf("a2:%v", err)
		return
	}
	atlas.Format, err = p.r.ReadString('\n')
	if err != nil {
		fmt.Printf("a3:%v", err)
		return
	}
	atlas.Filter, err = p.r.ReadString('\n')
	if err != nil {
		fmt.Printf("a4:%v", err)
		return
	}
	atlas.Repeat, err = p.r.ReadString('\n')
	if err != nil {
		fmt.Printf("a5:%v", err)
		return
	}

	return
}

func (p *Parser) Parse() ([]*Atlas, error) {
	_, err := p.r.ReadString('\n')
	if err != nil {
		fmt.Printf("e1:%v", err)
		return nil, err
	}
	var atlas []*Atlas
	a, err := p.NewAtlas()
	if err != nil {
		fmt.Printf("e1:%v", err)
		return nil, err
	}
	for {
		line, err := p.r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("e1:%v", err)
			return nil, err
		}

		if len(line) == 1 {
			atlas = append(atlas, a)
			a, err = p.NewAtlas()
			if err != nil {
				fmt.Printf("e1:%v", err)
				return nil, err
			}
			line, err := p.r.ReadString('\n')
			if err != nil {
				fmt.Printf("e1:%v", err)
				return nil, err
			}
			a.NewImage(p.r, line)
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

func main() {

	parser, err := NewParser("dragon-ess.atlas")
	if err != nil {
		panic(err)
	}
	defer parser.Close()
	atlas, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	for _, a := range atlas {
		err := a.Unpack()
		if err != nil {
			fmt.Printf("vim-go:%s:%v\n", a.Image, err)
		}
	}
}
