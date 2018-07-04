package image

import (
	"flag"
	"runtime"
	"fmt"
	"log"
	"github.com/PuerkitoBio/goquery"
	. "strings"
	"io/ioutil"
	"os"
	fp "path/filepath"
)

var (
	url = flag.String("url", "", "起始网址")
	downloadDir = flag.String("dir", "/Users/inke/download_pic", "自定义存放路径")
	sUrl = flag.String("furl", "", "自定义过滤网页链接的关键字")
	sPic = flag.String("fpic", "", "自定义过滤图片链接的关键字")
	sParent = flag.String("fparent", "", "自定义过滤图片父页面链接须包含的关键字")
	imgAttr = flag.String("img", "src", "自定义图片属性名称，如data-original")
	minSize = flag.Int("size", 150, "最小图片大小 单位kB")
 	maxNum = flag.Int("no", 20, "需要爬取的有效图片数量")
	recursive = flag.Bool("re", true, "是否需要递归当前页面链接")

	seen = History{m: map[string]bool{}}
	count = new(Counts)
	urlChan = make(chan *URL, 99999999)
	picChan = make(chan *URL, 99999999)
	done = make(chan int)

	goPicNum = make(chan int, 20)
	HOST string
)

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()

	if *url == "" {
		fmt.Println("Use -h or --help to get help!")
		return
	}

	fmt.Printf("Start:%v MinSize:%v MaxNum:%v Recursive:%v Dir:%v <img>attribution:%v\n",
		*url, *minSize, *maxNum, *recursive, *downloadDir, *imgAttr)
	fmt.Printf("Filter: URL:%v Pic:%v ParentPage:%v\n", *sUrl, *sPic, *sParent)

	u := NewURL(*url, nil, *downloadDir)
	HOST = u.Host
	fmt.Println(HOST)

	urlChan <- u
	seen.Add(u.Url)
	sleep(3)
	go HandleHTML()
	go HandlePic()

	<-done //等待信号，防止终端过早关闭
	log.Printf("图片统计：下载%v", count.Value("download"))
	log.Println("END")
}

func HandleHTML() {
	for {
		select {
		case u := <-urlChan:
			res := u.Get()
			if res == nil {
				log.Println("HTML response is nil! following process will not execute.")
				return
			}

			//goquery 会主动关闭res.Body
			doc, err := goquery.NewDocumentFromResponse(res)
			fuck(err)

			if *recursive {
				parseLinks(doc, u, urlChan, picChan)
			}
			parsePics(doc, u, picChan)

			count.Incr("page")
			log.Printf("当前爬取了 %v 个网页 %s", count.Value("page"), u.Url)
		default:
			log.Println("待爬取队列为空，爬取完成")
			done <- 1
		}
	}
	runtime.Gosched()
}

func HandlePic() {
	for u:= range picChan {
		goPicNum <- 1
		go func() {
			defer func() { <-goPicNum }()
			var data []byte
			res := u.Get()
			if res == nil {
				log.Println("HTML response is nil! following process will not execute.")
				return
			}

			defer res.Body.Close()
			count.Incr("pic")
			//if 200 <= res.StatusCode && res.StatusCode < 400
			if res.StatusCode == 200 {
				body := res.Body
				data, _ = ioutil.ReadAll(body)
				body.Close()
			} else {
				log.Println(res.StatusCode)
				sleep(3)
				return
			}

			if len(data) == 0 {
				return
			}

			if len(data) >= *minSize*1000 {
				cwd, e := os.Getwd()
				if e != nil {
					cwd = "."
				}
				picFile := fp.Join(cwd, u.FilePath)
				if exists(picFile) {
					return
				}

				picDir := fp.Dir(picFile)
				if !exists(picDir) {
					mkdirs(picDir)
				}

				f, e := os.Create(picFile)
				fatal(e)
				defer f.Close()
				_, e = f.Write(data)
				fatal(e)
				count.Incr("download")
				log.Printf("图片统计：下载%v 当前图片大小：%v kB", count.Value("download"), len(data)/1000)
			} else {
				log.Printf("爬取%v 当前图片大小：%v kB", count.Value("pic"), len(data)/1000)
			}
		}()
		runtime.Gosched() //显式地让出CPU时间给其他goroutine
	}
}

func parseLinks(doc *goquery.Document, parent *URL, urlChan, picChan chan *URL) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		url = Trim(url, " ")
		if ok {
			if HasPrefix(url, "#") || HasPrefix(ToLower(url), "javascript") || url == "" {
				return
			} else {
				new := NewURL(url, parent, *downloadDir)
				if seen.Has(new.Url) {
					log.Printf("链接已爬取，忽略 %v", new.Url)
				} else {
					seen.Add(new.Url)
					if (!IsPic(new.Url)) {
						if !Contains(new.Url, HOST) {
							log.Printf("链接已超出本站，忽略 %v", new.Url)
							return
						}
					}

					if !Contains(new.Path, *sUrl) {
						return
					}

					if IsPic(url) {
						picChan <- new
						log.Printf("New <a> PIC: %s", url)
					} else {
						select {
						case urlChan <- new:
							if Contains(url, "http") {
								log.Printf("New PAGE: %s", url)
							} else {
								log.Printf("New PAGE: %s --> %s", url, new.Url)
							}
						default:
							log.Println("url channel is full!!!!")
							sleep(3)
						}
					}
				}
			}
		}
	})
}

func parsePics(doc *goquery.Document, parent *URL, picChan chan *URL) {
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		url, ok := s.Attr(*imgAttr)
		url = Trim(url, " ")
		if ok {
			if HasPrefix(ToLower(url), "data") || url == "" {
				return
			} else {
				new := NewURL(url, parent, *downloadDir)
				if seen.Has(new.Url) {
					log.Printf("图片已爬取，忽略 %v", new.Url)
				} else {
					seen.Add(new.Url)
					if !Contains(parent.Path, *sParent) {
						log.Printf("父页面不满足过滤关键词，忽略 %v", new.Url)
						return
					}

					if !Contains(new.Path, *sPic) {
						log.Printf("不包含图片过滤关键词，忽略 %v", new.Url)
						return
					}

					if exists(new.FilePath) {
						log.Printf("图片已存在，忽略 %v", new.Url)
						return
					}

					picChan <- new
					log.Printf("New <img> PIC: %s", url)
				}
			}
		}
	})
}
