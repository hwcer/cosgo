package handler

import (
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

//getMethodName 获取接口名
func defPathParse(c *gin.Context) []string {
	strPath := strings.TrimRight(c.Request.URL.Path, "/")
	arrPath := strings.Split(strPath, "/")
	return  arrPath
}

func strFirstToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArray := []rune(str)
	strArray[0] = unicode.ToUpper(strArray[0])
	return string(strArray)
}
// joinPath("/a","b/","/c/","d") = /a/b/c/d
func joinPath(args ...string) string  {
	var nsp []string
	nsp = append(nsp,"")
	for _,k := range args{
		if k!=""{
			nsp = append(nsp,strings.Trim(k,"/"))
		}
	}
	return strings.Join(nsp,"/")
}


func TimeOut(d time.Duration,fn func()error) error {
	cher := make(chan error)

	go func() {
		cher <- fn()
	}()

	var err error
	select {
	case e := <-cher:
		err = e
	case <-time.After(d):
		err = nil
	}

	return err
}
