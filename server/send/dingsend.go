package send

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"dachshund-server/server/encrypt"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type DingDing struct {
	Secret  string
	Url     string
	Proxies string
}

type dingResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

//func (d DingDing) Send(msg string) error {
//	err := d.send(&textMessage{
//		MsgType: msgTypeText,
//		Text: textParams{
//			Content: msg,
//		},
//		At: atParams{
//			AtMobiles: []string{"13512910208"},
//			IsAtAll:   false,
//		},
//	})
//	if err != nil {
//		return err
//	}
//	return nil
//}

var pubKey []byte

func init() {
	pubKeyL, err := os.ReadFile("pubKey.pem")
	if err != nil {
		panic(err)
	}
	pubKey = pubKeyL
}

func (d DingDing) Send(info map[string]interface{}) error {
	if _, ok := info["Offset"]; ok {
		msg := fmt.Sprintf("#### Host:%s \n #### HostName:%s \n #### Error:%s \n #### offset:%s \n",
			info["IP"], info["HostName"], info["Error"], info["Offset"])
		//IP-timestamp-offset
		paraString := fmt.Sprintf("%s-%d-%s", info["IP"], time.Now().Unix(), info["Offset"])
		para := encrypt.RsaEncrypt([]byte(paraString), pubKey)
		return d.send(&actionCardMessage{
			MsgType: msgTypeActionCard,
			ActionCard: actionCardParams{
				Title:          "发生错误了",
				Text:           msg,
				SingleTitle:    "让我看看",
				SingleURL:      fmt.Sprintf("http://%s:4406/view/?parm=%s", info["IP"], url.QueryEscape(para)),
				BtnOrientation: "",
				HideAvatar:     "0",
			},
		})
	} else {
		var x string
		if i, ok := info["Last"].(int); ok {
			x = time.Unix(int64(i), 0).Format("2006-01-02 15:04:05")
		} else {
			x = "time error"
		}

		msg := fmt.Sprintf("#### 不好啦【%s】掉线了 \n #### 生命中的最后一秒钟是:%s \n ", info["IP"], x)
		return d.send(&markdownMessage{
			MsgType: msgTypeMarkdown,
			Markdown: markdownParams{
				Title: "掉线了",
				Text:  msg,
			},
			At: atParams{
				AtMobiles: nil,
				IsAtAll:   false,
			},
		})
	}

}

func (d DingDing) send(msg interface{}) error {
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	webURL := d.Url
	if len(d.Secret) != 0 {
		webURL += genSignedURL(d.Secret)
	}
	//resp, err := http.Post(webURL, "application/json", bytes.NewReader(m))
	urli := url.URL{}
	client := http.Client{}
	if d.Proxies != "" {
		urlproxy, _ := urli.Parse(d.Proxies)

		client = http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(urlproxy),
			},
		}
	}

	reqest, err := http.NewRequest(http.MethodPost, webURL, bytes.NewReader(m))
	if err != nil {
		return err
	}
	reqest.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dr dingResponse
	err = json.Unmarshal(data, &dr)
	if err != nil {
		return err
	}
	if dr.Errcode != 0 {
		return fmt.Errorf("dingrobot send failed: %v", dr.Errmsg)
	}

	return nil
}
func genSignedURL(secret string) string {
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	sign := fmt.Sprintf("%s\n%s", timeStr, secret)
	signData := computeHmacSha256(sign, secret)
	encodeURL := url.QueryEscape(signData)
	return fmt.Sprintf("&timestamp=%s&sign=%s", timeStr, encodeURL)
}
func computeHmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
