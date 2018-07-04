package image

import (
	"sync"
	. "strings"
	fp "path/filepath"
	"net/http"
	"math/rand"
	"time"
	"fmt"
)

type Counts struct {
	sync.RWMutex
	page, pic, download uint64
}

func (n *Counts) Incr(key string) {
	n.Lock()
	defer n.Unlock()
	switch key {
	case "page":
		n.page += 1
	case "pic":
		n.pic += 1
	case "download":
		n.download += 1
	}
}

func (n *Counts) Value(key string) uint64 {
	n.Lock()
	defer n.Unlock()
	switch key {
	case "page":
		return n.page
	case "pic":
		return n.pic
	case "download":
		return n.download
	default:
		return 0
	}
}

type History struct {
	m map[string]bool
	lock sync.RWMutex
}

func (h *History) Has(s string) bool {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.m[s]
}

func (h *History) Add(s string) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	h.m[s] = true
}

type URL struct {
	Url 		string
	Type 		string
	Protocol 	string
	Host 		string
	Port 		string
	Name 		string
	Origin 		string
	Path 		string

	Parent 		*URL
	FilePath 	string
}

func NewURL(url string, p *URL, dir string) *URL {
	u := new(URL)
	u.Url = url
	u.Parent = p
	u.Prepare(dir)
	return u
}

func (u *URL) Prepare(dir string) {
	if !HasPrefix(u.Url, "http") {
		if !HasPrefix(u.Url, "//") {
			u.Url = "//" + fp.Join(u.Parent.Host, u.Url)
			u.Url = Replace(u.Url, "\\", "/", -1)
		}
		u.Url = "http:" + u.Url
	}
	part := Split(u.Url, "/")
	u.Protocol = part[0]
	hp := Split(part[2], ":")

	u.Host = hp[0]
	if len(hp) > 1 {
		u.Port = hp[1]
	}

}

func (u *URL) Get() (res *http.Response) {
	req, _ := http.NewRequest("GET", u.Url, nil)
	req.Header.Set("User-Agent", ua.Random())
	if u.Parent != nil {
		req.Header.Set("Referer", u.Parent.Url)
	}
	client := &http.Client{Timeout: time.Second * 10}
	res, err := client.Do(req)

	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		fuck(err)
		sleep(3)
		return res
	}

	if res.StatusCode != 200 {
		fmt.Printf("Status code: %v", res.StatusCode)
		sleep(3)
	}

	return res
}

type UA struct {
	ua []string
}

var userAgent = []string{
	"Mozilla/5.0 (compatible, MSIE 10.0, Windows NT, DigExt)",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, 360SE)",
	"Mozilla/4.0 (compatible, MSIE 8.0, Windows NT 6.0, Trident/4.0)",
	"Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0,",
	"Opera/9.80 (Windows NT 6.1, U, en) Presto/2.8.131 Version/11.11",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, TencentTraveler 4.0)",
	"Mozilla/5.0 (Windows, U, Windows NT 6.1, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Macintosh, Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Macintosh, U, Intel Mac OS X 10_6_8, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Linux, U, Android 3.0, en-us, Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
	"Mozilla/5.0 (iPad, U, CPU OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, Trident/4.0, SE 2.X MetaSr 1.0, SE 2.X MetaSr 1.0, .NET CLR 2.0.50727, SE 2.X MetaSr 1.0)",
	"Mozilla/5.0 (iPhone, U, CPU iPhone OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"MQQBrowser/26 Mozilla/5.0 (Linux, U, Android 2.3.7, zh-cn, MB200 Build/GRJ22, CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1"}

var ua = UA{ua: userAgent}

func (u *UA) Random() string {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return userAgent[r.Intn(len(u.ua))]
}