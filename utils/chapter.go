package utils

import (
	"bytes"
	"fmt"
	"html"
	"io"
)

var (
	IsParagraph = false
	Blank       = []byte{}
)

type chapterContent struct {
	isNotSub bool
	Title    string
	Content  [][]byte
	Key      string
	Order    int
}

type chapterInfo struct {
	Title string
	Key   string
	Order int
}

type ChapterInfo struct {
	Title string
	Key   string
	Order int
	Sub   []*chapterInfo
}

func (c *chapterContent) Append(content []byte) {
	content = []byte(html.EscapeString(string(content)))
	c.Content = append(c.Content, content)

}

func (c *chapterContent) SetTitle(title string) {
	c.Title = title
}

func (c *chapterContent) Restore(title, key string, order int) {
	c.Title = title
	c.Key = key
	c.Order = order
	c.Content = make([][]byte, 0)
}

func (c *chapterContent) GetInfo() *chapterInfo {
	return &chapterInfo{
		Title: c.Title,
		Key:   c.Key,
		Order: c.Order,
	}
}

func (c *chapterContent) ToHtml() []byte {
	body := []byte{}
	if c.isNotSub {
		body = append(body, []byte(fmt.Sprintf("<a name=\"%s\"/><h1 id=\"%s\">%s</h1>\n", c.Key, c.Key, c.Title))...)
	} else {
		body = append(body, []byte(fmt.Sprintf("<a name=\"%s\"/><h2 id=\"%s\">%s</h2>\n", c.Key, c.Key, c.Title))...)
	}

	for _, l := range c.Content {
		body = append(body, generateHtmlP(l)...)
	}
	if len(c.Content) > 0 {
		body = append(body, []byte("<mbp:pagebreak/>\n")...)
	}
	return body

}

func generateHtmlP(line []byte) []byte {
	ps := []byte("<p>")
	pe := []byte("</p>\n")
	return bytes.Join([][]byte{ps, line, pe}, []byte(""))
}

type Chapter struct {
	content        chapterContent
	currentOrder   int
	subChapterList []chapterContent
	subLength      uint
}

func NewChapter(title string) Chapter {
	return Chapter{
		content: chapterContent{
			Title:    title,
			Content:  [][]byte{},
			Key:      "chap1",
			Order:    1,
			isNotSub: true,
		},
		subChapterList: make([]chapterContent, 0),
		currentOrder:   1,
	}
}

func (c *Chapter) SetOrder(i int) {
	c.currentOrder = i
}
func (c *Chapter) Restore(title string, order int) {
	c.subLength = 0
	c.content.Title = title
	c.content.Content = make([][]byte, 0)
	c.content.Order = order
	c.content.Key = fmt.Sprintf("chap%d", order)
	c.subChapterList = make([]chapterContent, 0)
	c.currentOrder = order
}

func (c *Chapter) AddSubChapter(title string) {
	c.currentOrder++
	c.subLength++
	subChapter := chapterContent{
		Title:   title,
		Content: make([][]byte, 0),
		Order:   c.currentOrder,
		Key:     fmt.Sprintf("chap%d", c.currentOrder),
	}
	c.subChapterList = append(c.subChapterList, subChapter)
}
func (c *Chapter) GetInfo() *ChapterInfo {
	info := c.content.GetInfo()
	maininfo := &ChapterInfo{
		Title: info.Title,
		Key:   info.Key, Order: info.Order,
		Sub: make([]*chapterInfo, c.subLength),
	}
	for i, sub := range c.subChapterList {
		maininfo.Sub[i] = sub.GetInfo()
	}
	return maininfo
}

func (c *Chapter) Append(content []byte) {
	if c.subLength < 1 {
		c.content.Append(content)
	} else {
		c.subChapterList[c.subLength-1].Append(content)
	}
}

func (c *Chapter) Flush(writer io.Writer) (nextIndex int, err error) {
	if _, err = writer.Write(c.content.ToHtml()); err != nil {
		return
	}
	for _, subchap := range c.subChapterList {
		if _, err = writer.Write(subchap.ToHtml()); err != nil {
			return
		}
	}
	return c.currentOrder + 1, nil
}
