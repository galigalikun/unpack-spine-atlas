package unpackspineatlas

import (
	"bufio"
	"fmt"
	"strings"
)

type Atlas struct {
	Image  string
	Size   string
	Format string
	Filter string
	Repeat string
	Images []*Image
}

func (a *Atlas) NewImage(r *bufio.Reader, line string) (err error) {
	if len(line) == 1 {
		return fmt.Errorf("image name error")
	}
	image := &Image{
		Name: fmt.Sprintf("%s.png", strings.Trim(line, "\n")),
	}
	err = image.parseRotate(r.ReadString('\n'))
	if err != nil {
		return err
	}
	err = image.parseXY(r.ReadString('\n'))
	if err != nil {
		return err
	}
	err = image.parseSize(r.ReadString('\n'))
	if err != nil {
		return err
	}
	err = image.parseOrig(r.ReadString('\n'))
	if err != nil {
		return err
	}
	err = image.parseOffset(r.ReadString('\n'))
	if err != nil {
		return err
	}
	err = image.parseIndex(r.ReadString('\n'))
	if err != nil {
		return err
	}

	a.Images = append(a.Images, image)

	return nil
}
