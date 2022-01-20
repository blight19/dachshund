package view

import (
	"dachshund-server/client/decrypt"
	"dachshund-server/utils"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	PrvKey  []byte
	IP      string
	LogPath string
}

func (s Server) check(IP string, thatTime string) bool {
	if s.IP != IP {
		return false
	}
	thatTimeT, err := strconv.Atoi(thatTime)
	if err != nil {
		return false
	}
	now := time.Now().Unix()
	// 5 hours
	if now-int64(thatTimeT) > 60*60*5 {
		return false
	}
	return true
}

func clip(num int, th int) int {
	if num > th {
		return num - th
	}
	return 0
}

func init() {
	file, _ := os.Create("web.log")
	gin.DefaultWriter = file
}

func (s Server) Serve() {
	router := gin.Default()

	gin.SetMode(gin.ReleaseMode)
	err := router.SetTrustedProxies([]string{"10.50.1.1/32", "10.50.1.2/31", "10.50.1.4/30",
		"10.50.1.8/29", "10.50.1.16/28", "10.50.1.32/27", "10.50.1.64/26",
		"10.50.1.128/25", "10.50.2.0/23", "10.50.4.0/22", "10.50.8.0/21", "10.50.16.0/20",
		"10.50.32.0/19", "10.50.64.0/18", "10.50.128.0/17"})
	if err != nil {
		return
	}
	//router.SetFuncMap(template.FuncMap{
	//	"safe": func(str string) template.HTML {
	//		return template.HTML(str)
	//	},
	//})
	//router.LoadHTMLGlob("templates/*")
	router.GET("/view", func(c *gin.Context) {

		parmb64 := c.Query("parm")
		parm, err := base64.StdEncoding.DecodeString(parmb64)
		parms, err := decrypt.RsaDecrypt(parm, s.PrvKey)
		//IP-timestamp-offset
		parmSlice := strings.Split(string(parms), "-")
		IP := parmSlice[0]
		sendTime := parmSlice[1]
		if !s.check(IP, sendTime) {
			c.String(http.StatusBadRequest, "请求已过期")
			return
		}
		offset := parmSlice[2]
		result := make([]byte, 4*1024)
		f, err := os.Open(s.LogPath)
		if err != nil && err != io.EOF {
			c.String(http.StatusInternalServerError, "文件找不到")
			return
		}
		defer f.Close()
		offsetInt, err := strconv.Atoi(offset)
		offsetInt = clip(offsetInt, 1024*2)
		readByte, err := f.ReadAt(result, int64(offsetInt))
		if err != nil && err != io.EOF {
			c.Status(http.StatusBadRequest)
			log.Error(err)
			return
		}

		resultStr := string(result[:readByte])
		c.String(http.StatusOK, resultStr)

		//resultStr = strings.Replace(resultStr, "\n", "<br />", -1)
		//c.HTML(http.StatusOK, "log.html", gin.H{"res": resultStr})

	})
	router.Run(":4406")
}
func NewServer(PrvKey []byte, LogPath string, NetCardName string) Server {
	ip, err := utils.LocalIpByName(NetCardName)
	if err != nil {
		panic(err)
	}
	return Server{
		PrvKey:  PrvKey,
		IP:      ip,
		LogPath: LogPath,
	}
}

//func main() {
//	prvKeyL, err := os.ReadFile("prvKey.pem")
//	if err != nil {
//		panic(err)
//	}
//	s := Server{
//		prvKey:  prvKeyL,
//		IP:      "10.50.101.50",
//		logPath: "F:\\MysqlMonitor\\MyIncrLog\\client\\1.log",
//	}
//	s.Serve()
//}
