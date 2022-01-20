package send

import (
	"testing"
)

func TestDingDing_Sign(t *testing.T) {
	ding := DingDing{
		Secret:  "SECcdbde934481db202cf7e3aed0c3df77fc4bc6923214c99041ebf94a9d5f3ad26",
		Url:     "https://oapi.dingtalk.com/robot/send?access_token=a2b1ab212c3817f1514c9d0cf4357b24f22c9f084b6971f4cfdeb33522ae2d09",
		Proxies: "",
	}

	err := ding.Send(map[string]interface{}{})
	if err != nil {
		t.Errorf("Send failed ,Error %s", err)
	}

}
