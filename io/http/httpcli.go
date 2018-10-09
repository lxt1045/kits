package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	DEF_DEAD_LINE = 3000 * time.Millisecond // 默认才http请求时间
	DEF_IDLE_CONN = 10                      // 默认的IDLE连接
)

var (
	clientPool *sync.Pool //加一个缓冲池，以保证高性能和线程安全
)

func init() {
	clientPool = &sync.Pool{
		New: newClient,
	}
}

// 默认Http的客户端，超时为DEF_DEAD_LINE， 写接口的时候确认是否需要这么小延时的接口
func newClient() interface{} {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, DEF_DEAD_LINE)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost:   DEF_IDLE_CONN,
			ResponseHeaderTimeout: DEF_DEAD_LINE,
		},
	}
}

// POST 根据Token 获取信息可以调用这个接口, 输出： out_resp
func PostTokenGetRsp(post_url string, token string, out_resp interface{}) error {
	req, err := http.NewRequest("POST", post_url, nil)
	req.Header.Set("token", token)
	req.Header.Set("Content-Type", "application/json")

	client := clientPool.Get().(*http.Client)
	defer clientPool.Put(client)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http.client.Do(req) err: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return fmt.Errorf("ioutil.ReadAll(resp.Body).%v", err)
	}

	return json.Unmarshal(body, out_resp)
}

// POST 根据userid 获取信息可以调用这个接口, 输出： out_resp
func PostUrlWithUserID(post_url string, userid uint64, out_resp interface{}) error {
	req, err := http.NewRequest("POST", post_url, nil)
	req.Header.Set("Userid", strconv.FormatUint(userid, 10))
	req.Header.Set("Content-Type", "application/json")

	client := clientPool.Get().(*http.Client)
	defer clientPool.Put(client)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http.client.Do(req) err: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return fmt.Errorf("ioutil.ReadAll(resp.Body).%v", err)
	}

	return json.Unmarshal(body, out_resp)
}

// POST 根据Token 获取信息可以调用这个接口, 输出： outResp
func PostReqForResp(url string, inReq interface{}, outResp interface{}) error {
	data, err := json.Marshal(inReq)
	if nil != err {
		return fmt.Errorf("json.marshal() err: %v", err)
	}

	reqbuf := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", url, reqbuf)
	req.Header.Set("Content-Type", "application/json")

	client := clientPool.Get().(*http.Client)
	defer clientPool.Put(client)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http.client.Do(req) err: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return fmt.Errorf("ioutil.ReadAll(resp.Body).%v", err)
	}

	return json.Unmarshal(body, outResp)
}

// 获取json数据，并将json数据解析为结构体
func GetUrlWithJsonData(qurl string, data interface{}) error {
	client := clientPool.Get().(*http.Client)
	defer clientPool.Put(client)

	resp, err := client.Get(qurl)
	if err != nil {
		return fmt.Errorf("http.Get(%s) err: %v", qurl, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return fmt.Errorf("ioutil.ReadAll(resp.Body) err: %v", err)
	}

	if err = json.Unmarshal(body, data); nil != err {
		return fmt.Errorf("json.Unmarshal(%s) err: %v", string(body), err)
	}

	return nil
}

// Token验证, 成功，返回userId， 失败标识Token错误
func TokenVerify(vurl string, data string, header map[string]string) (uint64, error) {
	return 0, nil
}
