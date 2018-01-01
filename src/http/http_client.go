/***********************************************************************
* @ http
* @ brief

* @ 通信技巧
	1、客户端20秒轮询一次服务端，问服务端有没有什么消息给我，比如双人聊天消息。
	2、如果取到消息，就把下一次轮训时间改短，比如5秒，再取到消息，继续改短，比如2秒，
	3、如果没消息就慢慢放长周期，比如：2秒，3秒，5秒，7秒，10秒，15秒，20秒
	4、直到有消息了，又再次把周期变短
	5、聊天模块的缩短程度，可以单独做大些

* @ author zhoumf
* @ date 2017-4-25
***********************************************************************/
package http

import (
	"bytes"
	"common"
	"common/net/meta"
	"gamelog"
	"net/http"
	"time"
)

// ------------------------------------------------------------
//! 底层接口，业务层一般用不到
func PostReq(url string, b []byte) []byte {
	ack, err := http.Post(url, "text/HTML", bytes.NewReader(b))
	if err == nil {
		backBuf := make([]byte, ack.ContentLength)
		ack.Body.Read(backBuf)
		ack.Body.Close()
		return backBuf
	} else {
		gamelog.Error("PostReq url: %s \r\nerr: %s \r\n", url, err.Error())
		return nil
	}
}

// 已验证：此函数失败，resp是nil，那resp.Body.Close()就不能无脑调了
// resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
// resp.Body.Close()

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string, meta *meta.Meta) {
	go _RegistToSvr(destAddr, meta)
}
func _RegistToSvr(destAddr string, meta *meta.Meta) {
	buf, _ := common.ToBytes(meta)
	for {
		http.DefaultClient.Timeout = 3 * time.Second
		if PostReq(destAddr+"reg_to_svr", buf) == nil {
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}
}
