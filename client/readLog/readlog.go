package readLog

import (
	"bufio"
	"dachshund-server/client/matchLog"
	"dachshund-server/client/send"
	"fmt"
	"io"
	"os"
	"time"
)

type LogReader struct {
	Sender    send.Sender
	LogPath   string
	StreamKey string
	HeartKey  string
	Matcher   matchLog.Matcher
	SkipOnce  bool
}

func sendTo(s send.Sender, out <-chan map[string]interface{}) {
	for {
		err := s.Send(<-out)
		if err != nil {
			return
		}
	}
}

func (l LogReader) Run() {

	var i int64 = 0
	in, out := l.Matcher.Match()
	// send message to
	go sendTo(l.Sender, out)
	// whether skip log when open at the first time
	if l.SkipOnce {
		ff, err := os.Open(l.LogPath)
		if err != nil {
			panic(err)
		}
		stat, _ := ff.Stat()
		i = stat.Size()
	}

	for {
		f, err := os.Open(l.LogPath)
		finfo, err := f.Stat()
		if err != nil {
			fmt.Println(err)
		}
		if finfo.Size() < i {
			i = 0
		}
		seek, err := f.Seek(i, 0)
		if err != nil {
			fmt.Println(seek, err)
		}
		rd := bufio.NewReader(f)
		//send the new line to matcher chan
		for {
			flag := 0
			line, err := rd.ReadString('\n')
			if err != nil || io.EOF == err {
				flag = 1
				if io.EOF != err {
					fmt.Println(err)
				}
			}
			in <- map[string]interface{}{"Error": line, "Offset": i}
			n := len([]byte(line))
			i = i + int64(n)
			if flag == 1 {
				break
			}

		}
		time.Sleep(time.Second)
		f.Close()
	}

}
