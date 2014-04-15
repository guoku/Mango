package crawler

import (
    "Mango/gojobs/log"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "regexp"
    "strings"
    "time"
)

//返回值font是商品的页面，detail是商品的详情页面，instock表示下架与否
//这里下架与否的判断设计得比较不好,如果抓取正常，instock是未知的，只有进行解析后才知道结果
//而如果出现了err，则应该看instock与否，如果下架了，这个itemid就不需要保存了
//这个是对Fetch的封装，因为返回的错误类型需要用来判断是否要保存这个item
func FetchItem(itemid string, shoptype string) (font, detail string, instock bool, err error, isWeb bool) {
    log.Infof("start to fetch %s", itemid)
    font, detail, err, isWeb = Fetch(itemid, shoptype)
    if err != nil {
        log.ErrorfType(FETCH_ERR, "%s failed", itemid)
        if err.Error() != "404" {
            //说明不是因为商品下架而导致的失败
            instock = true
            return
        } else {
            //商品的页面已经找不到了
            instock = false
            return
        }
    }
    log.Infof("%s fetched successed!", itemid)
    return
}

//这种方式在天猫商品里会多一次访问，所以建议少用
func FetchWithOutType(itemid string) (html, detail, shoptype string, instock bool, err error) {
    shoplink := fmt.Sprintf("http://a.m.taobao.com/i%s.htm", itemid)
    instock = true
    transport := getTransport()
    client := &http.Client{Transport: transport}
    req, err := http.NewRequest("GET", shoplink, nil)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    useragent := userAgentGen()
    req.Header.Set("User-Agent", useragent)
    resp, err := client.Do(req)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    if resp == nil {
        err = errors.New("response is nil")
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    defer func() {
        if resp != nil {
            resp.Body.Close()
        }
    }()
    if resp.StatusCode == 200 {
        resplink := resp.Request.URL.String()
        if strings.Contains(resplink, "h5") {
            err = errors.New("taobao forbidden")
            log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
            return
        }
        if resplink != shoplink {
            shoptype = "tmall.com"
        } else {
            shoptype = "taobao.com"
        }
        bytedata, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
            return "", "", "", true, err
        }
        html = string(bytedata)

    } else {
        instock = false
        err = errors.New("404")
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    resp.Body.Close()
    detailurl := ""
    if shoptype == "taobao.com" {
        detailurl = fmt.Sprintf("http://a.m.taobao.com/da%s.htm", itemid)
    } else {
        detailurl = fmt.Sprintf("http://a.m.tmall.com/da%s.htm", itemid)
    }

    req, err = http.NewRequest("GET", detailurl, nil)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    req.Header.Set("User-Agent", useragent)
    resp, err = client.Do(req)
    if err != nil || resp == nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    if resp.StatusCode == 200 {
        resplink := resp.Request.URL.String()
        if strings.Contains(resplink, "h5") {
            err = errors.New("taobao forbidden")
            log.Error(err)
            return
        }
        bytedata, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
            return "", "", "", true, err
        }
        detail = string(bytedata)
    } else {
        log.Info(resp.StatusCode)
        err = errors.New(resp.Status)
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        instock = false
        return
    }

    resp.Body.Close()
    re := regexp.MustCompile("\\<style[\\S\\s]+?\\</style\\>")
    re2 := regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
    html = re.ReplaceAllString(html, "")
    detail = re.ReplaceAllString(detail, "")
    html = re2.ReplaceAllString(html, "")
    detail = re2.ReplaceAllString(detail, "")
    err = nil
    return
}

//根据商品id和店铺类型抓取页面
func Fetch(itemid string, shoptype string) (html string, detail string, err error, isWeb bool) {
    url := ""
    detailurl := ""
    if shoptype == "tmall.com" {
        url = "http://a.m.tmall.com/i" + itemid + ".htm"
        detailurl = "http://a.m.tmall.com/da" + itemid + ".htm"
    } else {
        url = "http://a.m.taobao.com/i" + itemid + ".htm"
        detailurl = "http://a.m.taobao.com/da" + itemid + ".htm"
    }
    transport := getTransport()
    client := &http.Client{Transport: transport}
    req, err := http.NewRequest("GET", url, nil)
    useragent := userAgentGen()
    req.Header.Set("User-Agent", useragent)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    log.Info("start to do request")
    resp, err := client.Do(req)
    log.Info("request has been done")
    if err != nil {
        if resp == nil {
            log.Info("当proxy不可达时，resp为空")
        }
        time.Sleep(1 * time.Second)
        log.ErrorfType(HTTP_ERR, "商品 %s is %s", itemid, err)
        return
    }
    defer resp.Body.Close()
    resplink := resp.Request.URL.String()
    log.Info(resplink)
    if strings.Contains(resplink, "cloud-jump") {
        html, detail, err = FetchWeb(itemid, shoptype)
        isWeb = true
        return
    }
    if resp.StatusCode == 200 {
        //fmt.Println(resp.Request.URL.String())
        resplink := resp.Request.URL.String()
        if strings.Contains(resplink, "h5") {
            html = ""
            detail = ""
            err = errors.New("taobao forbidden")
            log.Error("taobao forbidden")
            return
        }
        robots, e := ioutil.ReadAll(resp.Body)
        if e != nil {
            err = e
            log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
            return
        }
        html = string(robots)
    } else {
        err = errors.New("404")
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    resp.Body.Close()
    req, err = http.NewRequest("GET", detailurl, nil)
    req.Header.Set("User-Agent", useragent)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    resp, err = client.Do(req)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    if resp.StatusCode == 200 {
        //fmt.Println(resp.Request.URL.String())
        resplink := resp.Request.URL.String()
        if strings.Contains(resplink, "h5") {
            html = ""
            detail = ""
            err = errors.New("taobao forbidden")
            log.Error("taobao forbidden")
            return
        }
        robots, e := ioutil.ReadAll(resp.Body)
        if e != nil {
            err = e
            log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
            return
        }
        detail = string(robots)
    } else {
        html = ""
        err = errors.New(resp.Status)
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return
    }
    resp.Body.Close()
    re := regexp.MustCompile("\\<style[\\S\\s]+?\\</style\\>")
    re2 := regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
    html = re.ReplaceAllString(html, "")
    detail = re.ReplaceAllString(detail, "")
    html = re2.ReplaceAllString(html, "")
    detail = re2.ReplaceAllString(detail, "")
    err = nil
    return
}

func FetchWeb(itemid string, shoptype string) (string, string, error) {
    var fonturl string
    if shoptype == "tmall.com" {
        fonturl = fmt.Sprintf("http://detail.tmall.com/item.htm?id=%s", itemid)
    } else {
        fonturl = fmt.Sprintf("http://item.taobao.com/item.htm?id=%s", itemid)
    }
    transport := getTransport()
    client := &http.Client{Transport: transport}
    req, _ := http.NewRequest("GET", fonturl, nil)
    useragent := userAgentGen()
    req.Header.Set("User-Agent", useragent)
    req.Header.Set("Cookie", "cna=I2H3CtFnDlgCAbRP3eN/4Ujy; t=2609558ec16b631c4a25eae0aad3e2dc; w_sec_step=step_login; x=e%3D1%26p%3D*%26s%3D0%26c%3D0%26f%3D0%26g%3D0%26t%3D0%26__ll%3D-1%26_ato%3D0; lzstat_uv=26261492702291067296|2341454@2511607@2938535@2581747@3284827@2581759@2938538@2817407@2879138@3010391; tg=0; _cc_=URm48syIZQ%3D%3D; tracknick=; uc3=nk2=&id2=&lg2=; __utma=6906807.613088467.1388062461.1388062461.1388062461.1; __utmz=6906807.1388062461.1.1.utmcsr=(direct)|utmccn=(direct)|utmcmd=(none); mt=ci=0_0&cyk=0_0; _m_h5_tk=6457881dd2bbeba22fc0b9d54ec0f4d9_1389601777274; _m_h5_tk_enc=3c432a80ff4e2f677c6e7b8ee62bdb48; _tb_token_=uHyzMrqWeUaM; cookie2=3f01e7e62c8f3a311a6f83fb1b3456ee; wud=wud; lzstat_ss=2446520129_1_1389711010_2581747|2258142779_0_1389706922_2938535|1182737663_4_1389706953_3284827|942709971_0_1389706966_2938538|2696785043_0_1389707052_2817407|50754089_2_1389707124_2879138|2574845227_1_1389707111_3010391|377674404_1_1389711010_2581759; linezing_session=3lJ2NagSIjQvEYbpCk5o8clc_1389693042774lS4I_5; swfstore=254259; whl=-1%260%260%261389692419141; ck1=; uc1=cookie14=UoLU4ni6x8i9JA%3D%3D; v=0")
    resp, err := client.Do(req)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return "", "", err
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.ErrorfType(HTTP_ERR, "%s is %s", itemid, err)
        return "", "", err
    }
    fonthtml := string(body)

    return fonthtml, "", nil

}
