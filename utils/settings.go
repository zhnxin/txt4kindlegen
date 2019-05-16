package utils

import (
	"fmt"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	BlankLineRepx = regexp.MustCompile(`^\s*$`)
)

func init() {
}

type Config struct {
	Title           string
	Cover           string
	Author          string
	Chapter         string
	SubChapter      string
	Encoding        string
	File            string
	ChapterRegex    *regexp.Regexp
	SubChapterRegex *regexp.Regexp
	Compress        bool
	decode          *encoding.Decoder
}

func NewConfig(title, cover, author, chapter, subchapter, encoding, file string) *Config {
	config := &Config{
		Title:           title,
		Cover:           cover,
		Author:          author,
		Chapter:         chapter,
		SubChapter:      subchapter,
		Encoding:        encoding,
		File:            file,
		ChapterRegex:    nil,
		SubChapterRegex: nil,
		decode:          nil,
	}
	return config
}
func (c *Config) Update(file, title, author, cover string) {
	if file != "" {
		c.File = file
	}
	if title != "" {
		c.Title = title
	}
	if author != "" {
		c.Author = author
	}
	if cover != "" {
		c.Cover = cover
	}
}

func (config *Config) Check() (err error) {
	switch config.Encoding {
	case "GB18030", "gb18030":
		config.decode = simplifiedchinese.GB18030.NewDecoder()
	case "GBK", "gbk":
		config.decode = simplifiedchinese.GBK.NewDecoder()
	case "UTF8", "utf8", "utf-8", "":
		config.decode = nil
	default:
		return fmt.Errorf("Unsupport encoding[GB18030,GBK,UTF8(default)]:%s", config.Encoding)
	}
	if _, err = os.Stat(config.File); os.IsNotExist(err) {
		return
	}
	config.ChapterRegex, err = regexp.Compile(config.Chapter)
	if err == nil && config.SubChapter != "" {
		config.SubChapterRegex, err = regexp.Compile(config.SubChapter)
	}
	return
}

func NewConfigWithFile(configFile string) (config *Config, err error) {
	config = &Config{}
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		return
	}
	return
}

func (c *Config) Decode(content []byte) ([]byte, error) {
	if c.decode != nil {
		return c.decode.Bytes(content)
	}
	return content, nil
}
