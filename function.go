package image

import (
	"log"
	"fmt"
	"time"
	"os"
	. "strings"
)

func fuck(e error) {
	if (e != nil) {
		log.Println(fmt.Printf("fucking: %v", e))
		sleep(3)
	}
}

func sleep(n int) {
	time.Sleep(time.Second * time.Duration(n))
}

func fatal(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func trace(msg string) func() {
	start := time.Now()
	return func() {
		log.Printf("exit %s (%s)", msg, time.Since(start))
	}
}

func mkdirs(s string) {
	err := os.MkdirAll(s, 0777)
	if err != nil {
		fatal(err)
	} else {
		fmt.Printf("Create Directory OK!")
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func IsPic(url string) bool {
	exts := []string{"jpg", "png", "gif", "csv"}
	for _, ext := range exts {
		if Contains(ToLower(url), "." + ext) {
			return true
		}
	}
	return false
}

func print(s interface{}) {
	fmt.Printf("%v\n", s)
}