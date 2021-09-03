package unpackspineatlas

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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
	// return image.Rect(500, 600, 100, 200) // 100, 200, 400, 400
	if i.rotate {
		return image.Rect(i.xy.X, i.xy.Y, i.xy.X+i.size.Y, i.xy.Y+i.size.X)
	}
	return image.Rect(i.xy.X, i.xy.Y, i.xy.X+i.size.X, i.xy.Y+i.size.Y)
}

func rotate(m *image.NRGBA) *image.NRGBA {
	rotate := image.NewNRGBA(image.Rect(0, 0, m.Bounds().Dy(), m.Bounds().Dx()))
	// 矩阵旋转
	for x := m.Bounds().Min.Y; x < m.Bounds().Max.Y; x++ {
		for y := m.Bounds().Max.X - 1; y >= m.Bounds().Min.X; y-- {
			//  设置像素点
			rotate.Set(m.Bounds().Max.Y-x, y-m.Bounds().Min.X, m.At(y, x))
		}
	}
	return rotate
}

func rotateImg(m image.Image) image.Image {
	rotateImg := image.NewNRGBA(image.Rect(0, 0, m.Bounds().Dy(), m.Bounds().Dx()))
	// 矩阵旋转
	for x := m.Bounds().Min.Y; x < m.Bounds().Max.Y; x++ {
		for y := m.Bounds().Max.X - 1; y >= m.Bounds().Min.X; y-- {
			//  设置像素点
			rotateImg.Set(m.Bounds().Max.Y-x, y-m.Bounds().Min.X, m.At(y, x))
		}
	}
	return rotateImg
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
	log.Printf("rotate->%s", ss)
	if strings.Index(ss, "true") != -1 {
		log.Println("rotate true!!")
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
	y, e := strconv.Atoi(strings.TrimSpace(point[1]))
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
func (a *Atlas) Unpack(outpath string) error {
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
	for i := 0; i < len(a.Images); i++ {
		go func(img *Image) {
			imgRect := img.Rect()
			log.Printf("debug rect:%v", imgRect)
			imageFile := filepath.Join(outpath, img.Name)
			newImg := atlasImage.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(imgRect)

			// 获取origin大小
			var orignRect image.Rectangle
			orignRect = image.Rect(0, 0, img.orig.X, img.orig.Y)

			if img.rotate {
				newImg = rotateImg(newImg)
			}

			// 根据origin大小创建新的图片
			rgba := image.NewNRGBA(orignRect)
			// 将图片内容写入到新图片中
			// fmt.Printf("offset: %d %d              ", img.offset.X, img.offset.Y)
			// fmt.Printf("bounds: %v\n", newImg.Bounds())
			// fmt.Printf("origin rect: %v\n", img.OrigRect())
			var rgbaRect image.Rectangle
			offsetX := img.offset.X
			if img.rotate {
				offsetY := img.orig.Y - img.offset.Y - img.size.Y
				rgbaRect = image.Rect(offsetX, offsetY, offsetX+img.size.X, offsetY+img.size.Y)
			} else {
				offsetY := img.orig.Y - img.offset.Y - img.size.Y
				rgbaRect = image.Rect(offsetX, offsetY, offsetX+img.size.X, offsetY+img.size.Y)
			}

			draw.Draw(rgba, rgbaRect, newImg, newImg.Bounds().Min, draw.Src)
			os.MkdirAll(filepath.Dir(imageFile), 0755)
			out, err := os.Create(imageFile)
			if err != nil {
				createImage <- fmt.Errorf("create dir error")
				return
			}
			err = png.Encode(out, rgba)
			if err != nil {
				createImage <- fmt.Errorf("crate png error")
				return
			}
			createImage <- nil
		}(a.Images[i])
	}

	for i := 0; i < len(a.Images); i++ {
		<-createImage
	}

	return nil
}
