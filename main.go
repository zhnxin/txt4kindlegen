package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"runtime"
	textTemplate "text/template"
	"time"

	"./assets"
	"./utils"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ConfigFile  = kingpin.Flag("config", "config file").Default("config.toml").Short('c').String()
	isInit      = kingpin.Flag("init", "make a config file").Short('i').Bool()
	isNotDelete = kingpin.Flag("delete", "is not to delete the temp file").Short('d').Bool()
	bookTemp    *template.Template
	ncxTemp     *textTemplate.Template
	opfTemp     *textTemplate.Template
	tocTemp     *textTemplate.Template
	CONFIG      = &utils.Config{}
)

func kindlegenSh() error {
	var cmd string = "kindlegen"
	if _, err := os.Stat(cmd); !os.IsNotExist(err) {
		cmd = "./" + cmd
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, cmd, "-dont_append_source", "-c1", "-o", CONFIG.Title+".mobi", "book.opf")
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
}

func kinlegenCmd() error {
	var cmd string = "kindlegen.exe"
	if _, err := os.Stat(cmd); !os.IsNotExist(err) {
		cmd = "./" + cmd
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, cmd, "-dont_append_source", "-c1", "-o", CONFIG.Title+".mobi", "book.opf")
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
}

func init() {
	if body, err := assets.Asset("assets/index.html"); err != nil {
		panic(err)
	} else if bookTemp, err = template.New("bookTemp").Parse(string(body)); err != nil {
		panic(err)
	}
	if body, err := assets.Asset("assets/toc.ncx"); err != nil {
		panic(err)
	} else if ncxTemp, err = textTemplate.New("ncxTemp").Parse(string(body)); err != nil {
		panic(err)
	}
	if body, err := assets.Asset("assets/toc.xhtml"); err != nil {
		panic(err)
	} else if tocTemp, err = textTemplate.New("tocTemp").Parse(string(body)); err != nil {
		panic(err)
	}
	if body, err := assets.Asset("assets/book.opf"); err != nil {
		panic(err)
	} else if opfTemp, err = textTemplate.New("opfTemp").Parse(string(body)); err != nil {
		panic(err)
	}
}

type ChapterTitles struct {
	Title    string
	Chapters []string
}

func main() {
	kingpin.Parse()
	if *isInit {
		conf, err := os.OpenFile("config_example.toml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Fatal(err)
		}
		body, err := assets.Asset("assets/config_example.toml")
		if err != nil {
			panic(err)
		}
		_, err = conf.Write(body)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	var err error
	CONFIG, err = utils.NewConfigWithFile(*ConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(CONFIG.Lang)
	if err = CONFIG.Check(); err != nil {
		log.Fatal(err)
	}
	output, err := os.OpenFile("index.html", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()
	if err = bookTemp.Execute(output, CONFIG); err != nil {
		log.Fatal(err)
	}
	if opf, err := os.OpenFile("book.opf", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755); err != nil {
		log.Fatal(err)
	} else {
		if err = opfTemp.Execute(opf, CONFIG); err != nil {
			opf.Close()
			log.Fatal(err)
		}
		opf.Close()
	}
	file, err := os.Open(CONFIG.File)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var line []byte
	scanner := bufio.NewScanner(file)
	chapter := utils.NewChapter(CONFIG.Title)
	chapinfo := struct {
		Chapters []*utils.ChapterInfo
	}{
		Chapters: []*utils.ChapterInfo{},
	}

	for scanner.Scan() {
		line, err = CONFIG.Decode(scanner.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		if utils.BlankLineRepx.Match(line) {
			continue
		}
		if CONFIG.ChapterRegex.Match(line) {
			nextOrder, err := chapter.Flush(output)
			if err != nil {
				log.Fatal("write body:", err)
			}
			chapinfo.Chapters = append(chapinfo.Chapters, chapter.GetInfo())
			chapter.Restore(string(line), nextOrder)
		} else if CONFIG.SubChapterRegex != nil && CONFIG.SubChapterRegex.Match(line) {
			chapter.AddSubChapter(string(line))
		} else {
			chapter.Append(line)
		}
	}
	_, err = chapter.Flush(output)
	if err != nil {
		log.Fatal(err)
	}
	chapinfo.Chapters = append(chapinfo.Chapters, chapter.GetInfo())
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
	file.Close()
	if _, err = fmt.Fprintln(output, "\n</body>\n</html>"); err != nil {
		log.Fatal(err)
	}
	output.Close()
	if toc, err := os.OpenFile("toc.ncx", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755); err != nil {
		log.Fatal(err)
	} else {
		if err = ncxTemp.Execute(toc, chapinfo); err != nil {
			toc.Close()
			log.Fatal(err)
		}
		toc.Close()
	}
	if toc, err := os.OpenFile("toc.xhtml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755); err != nil {
		log.Fatal(err)
	} else {
		if err = tocTemp.Execute(toc, chapinfo); err != nil {
			toc.Close()
			log.Fatal(err)
		}
		toc.Close()
	}
	if runtime.GOOS != "windows" {
		if err = kindlegenSh(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err = kinlegenCmd(); err != nil {
			log.Fatal(err)
		}
	}
	if !*isNotDelete {
		os.Remove("book.opf")
		os.Remove("index.html")
		os.Remove("toc.xhtml")
		os.Remove("toc.ncx")
	}
}
