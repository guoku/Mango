package crawler

import (
    "bytes"
    "compress/zlib"
    "container/ring"
    "github.com/qiniu/log"
    "math/rand"
    "net/http"
    "net/url"
    "time"
)

var pxyring *ring.Ring

func init() {
    length := len(proxys)
    pxyring = ring.New(length)
    for i := 0; i < length; i++ {
        pxyring.Value = proxys[i]
        pxyring = pxyring.Next()
    }
}

func getTransport() (transport *http.Transport) {
    /*
    	length := len(proxys)
    	r := rand.New(rand.NewSource(time.Now().UnixNano()))
    	proxy := proxys[r.Intn(length)]
    */
    pxyring = pxyring.Next()
    proxy := pxyring.Value.(string)
    log.Info("使用的proxy为：", proxy)
    url_i := url.URL{}
    url_proxy, _ := url_i.Parse(proxy)
    transport = &http.Transport{Proxy: http.ProxyURL(url_proxy), ResponseHeaderTimeout: time.Duration(30) * time.Second, DisableKeepAlives: true}
    return
}
func IsTmall(itemid string) (bool, error) {
    url := "http://a.m.taobao.com/i" + itemid + ".htm"
    request, _ := http.NewRequest("GET", url, nil)
    //transport := getTransport()
    client := &http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        return false, err
    } else {
        finalURL := resp.Request.URL.String()
        if finalURL == url {
            return false, nil
        } else {
            return true, nil
        }
    }
    resp.Body.Close()
    return true, nil
}

var UserAgents []string = []string{
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0",
    "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:22.0) Gecko/20100101 Firefox/22.0",
}

func userAgentGen() string {
    length := len(UserAgents)
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    return UserAgents[r.Intn(length)]
}
func Compress(data string) string {
    var b bytes.Buffer
    w := zlib.NewWriter(&b)
    w.Write([]byte(data))
    w.Close()
    return string(b.Bytes())
}

func Decompress(data string) (string, error) {
    buff := []byte(data)
    b := bytes.NewBuffer(buff)
    r, err := zlib.NewReader(b)
    if err != nil {
        log.Error(err)
        return "", err
    }
    rbuf := new(bytes.Buffer)
    rbuf.ReadFrom(r)
    return string(rbuf.Bytes()), nil
}
