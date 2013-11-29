#encoding=utf-8

from bs4 import BeautifulSoup
import re
from urlparse import parse_qs
from urlparse import urlparse
import pymongo
import time
import urllib2
import urllib 
import json
import thread
import codecs

states = set([u"北京",u"上海",u"天津",u"重庆",u"广东",u"江苏",u"浙江",u"山东",u"河北",u"山西",u"辽宁"
    ,u"吉林",u"河南",u"安徽",u"福建",u"江西",u"黑龙江",u"湖南",u"湖北",u"海南",u"四川",u"贵州",u"云南",
    u"陕西",u"甘肃",u"青海",u"台湾",u"西藏",u"内蒙古",u"广西",u"宁夏",u"新疆",u"香港",u"澳门",u"海外"])

def fetch(html):
    soup = BeautifulSoup(html)
    title = soup.title.text
    if title.find(u'商品屏蔽')>=0:
        return None,'error'
    desc = soup.title.string[0:-9]

    cattag = soup.p
    if cattag==None:
        print "已经下架了"
        return None,'error'
    atag = cattag.findChildren('a')
    cidtag = ''
    cidtag=atag[-1]
    cidurl = cidtag.attrs['href']
    cid = int(re.findall(r'\d+$',cidurl)[0])

    firstcat = atag[1].text
    secondcat = cidtag.text

    category = [firstcat,secondcat]

    details = soup.find_all('div', { 'class' : 'detail' })
        
    imgurls = []      
    imgtag = details[0]
    src = ''
    if imgtag.img != None:
        #有可能连首页图片都没有
        src = imgtag.img['src']
        src = re.sub('_\d+x\d+\.jpg','',src)
        imgurls.append(src)
        
    tables = soup.find_all('table')
    imgtable = tables[1]
    imgtags = imgtable.findChildren('img')
    for tag in imgtags:
        imgurl = tag['src']
        #imgurl = imgurl.replace('_70x70.jpg','')
        imgurl = re.sub('_\d+x\d+\.jpg','',imgurl)
        imgurls.append(imgurl)

            
    detail = details[1]
    instock = True
    if len(detail.findChildren('table'))==1:
       instock = False 
    judge = re.findall(ur'格：',detail.p.text)
    hasprom = True #默认都有促销
    secondhand = False
    if len(judge)>0:
        hasprom = False
    else:
        judge = re.findall(ur'价',detail.p.text)
        if len(judge)>0:
            
            secondhand = True
    promprice = 0
    if hasprom:
        prom = detail.p.strong
        if prom != None:
            tmp = re.findall("\d+.\d+",prom.text)
            if len(tmp)>0:
                promprice = float(tmp[0])
            

    p = detail.findChildren('p')
    startindex = 1
    
    if hasprom==False:
        startindex = 0
    pricetag = p[startindex].text
    price = 0
    if pricetag!=None:
        prices =re.findall(r'\d+\.\d+',pricetag)
        if len(prices)>0:
            price = float(prices[0])

    counttag = p[startindex+2].text
    tmp = re.findall(r'\d+',counttag)
    salecount = 0
    if len(tmp)>0:
        salecount = int(tmp[0])


    loctag = p[startindex+3].text
    state = ''
    city = ''
    if loctag.find(u'地')>0:
        location = loctag.split(u'：')[1].strip()
        if len(location)==2 or len(location)==3:
            state = location
            city = location
        elif len(location)>3:
            if location[0:2] in states :
                state = location[0:2]
                city = location[2:]
            elif location[0:3] in states:
                state = location[0:3]
                city = location[3:]

    #有些二手转让页面结构与普通商品不一样，有些又是一样的，所以需要特殊处理
    fix_btntag = soup.findChildren('td','link_btn fix_btn')
    deta = fix_btntag[0]
    detaurl = deta.a['href']
    itemid = re.findall('\d+',detaurl)[0]
    reviewtag = fix_btntag[1]
    reviews = 0
    if secondhand:
        rew = re.findall(r'\d+',reviewtag.a.text)
        if len(rew)>0:
            reviews = int(rew[0])
    else:
        rew = re.findall(r'\d+',reviewtag.a.span.text)
        if len(rew)>0:
            reviews = int(rew[0])
    nick = '' 
    for nametag in soup.select('body div.bd div.box div.detail p a img'):
        try:
            nameurl = nametag['src']
            o = urlparse(nameurl)
            nick = parse_qs(o.query)['nick'][0]
            nick= nick
            break
        except:
            pass

    result = {
            "desc":desc,
            "cid":cid,
            "promprice":promprice , #促销价格
            "price":price ,
            "category":category,
            "imgs":imgurls,
            "count":salecount,
            "state":state,
            "city":city,
            "reviews":reviews,
            "nick":nick,
            "instock":instock,
            "itemid":itemid
            } 
    return result,'success'


def fetchdetail(html):
    soup = BeautifulSoup(html)

    boxs = soup.find_all('div',{'class':'box'})
    reviewtag = ''
    noattr = False
    try:
        reviewtag = boxs[2].p.a.strong.text
    except:
        reviewtag = boxs[1].p.a.strong.text
        noattr = True 
    reviewregx = re.findall('\d+',reviewtag)
    reviews = 0
    if len(reviewregx)>0:
        reviews = int(reviewregx[0])
    if noattr:
        return {"reviews":reviews}
    attribox = boxs[3]
    table = attribox.div.table 
    trs = table.findChildren('tr')
    attri = {}
    for tr in trs:
        tds = tr.findChildren('td')
        key = tds[0].text.strip()[0:-1]
        value = tds[1].text.strip()
        attri[key] = value

    attri["reviews"]=reviews

    return attri


host='10.0.1.23'

def removeItem(pages,itemid):
    if len(itemid)>0:
        #防止空参数把所有数据都删除掉了
        pages.remove(itemid)

def updateItem(pages,item):
    pages.update({"itemid":item['itemid']},{'$set':item})
def process(pages,item):
    if item==None:
        time.sleep(10)
        return
    result = {}
    print '错误的itemid:', item['itemid']
    fontdata,statu = fetch(item['fontpage'])
    if statu=='error':
        #说明这个商品已经被屏蔽掉了，抓取下来的页面是错误页面，应该删除掉
        item['instock']=False
        item['parsed']=True
        item['updatetime'] = int(time.time()) 
        updateItem(pages,item)
        conn.close()
        print('这个商品被屏蔽了')
        return
    detaildata={}
    detaildata = fetchdetail(item['detailpage'])
    if detaildata['reviews']>0:
        fontdata['reviews']=detaildata['reviews']
    detaildata.pop('reviews')
    if item['shoptype']=='tmall.com':
        result['detail_url']='http://detail.tmall.com/item.htm?id='+item['itemid']
    else:
        result['detail_url'] = 'http://item.taobao.com/item.htm?id='+item['itemid']
    print "正确的itemid:",fontdata['itemid']
    result['num_iid'] =int(fontdata['itemid'])
    result['title'] = fontdata['desc']
    print result['title']
    result['nick'] = fontdata['nick']
    result['desc'] = fontdata['desc']
    result['cid'] = int(fontdata['cid'])
    result['sid'] = int(item['shopid'])
    result['location'] = {"state":fontdata["state"],"city":fontdata['city']}
    print fontdata['price']
    result['price'] = fontdata['price']
    result['promotion_price']=fontdata['promprice']
    result['item_imgs'] = fontdata['imgs'] #数组
    result['shop_type'] = item['shoptype']
    result['reviews_count'] = fontdata['reviews']
    result['monthly_sales_volume'] = fontdata['count']
    result['props']= detaildata
    result['in_stock'] = fontdata['instock']
    result['data_updated_time'] = item['updatetime']

    posturl = 'http://10.0.1.23:8080/scheduler/api/send_item_detail?token=d61995660774083ccb8b533024f9b8bb'
    js = json.dumps(result)
    #print js
    f = urllib2.urlopen(posturl,js)
    resp = f.read()
    print resp 
    statu = json.loads(resp)
    if statu['status']!='succeeded':
        print 'http request error'
        return
    item = {}
    item['instock']=fontdata['instock']
    item['parsed']=True
    item['updatetime'] = int(time.time()) 
    item['itemid']=fontdata['itemid']
    
    updateItem(pages,item)
    #conn.close()
    return result

if __name__=='__main__':
    conn=pymongo.Connection(host)
    db = conn['zerg']
    pages=db['pages']
    cur = pages.find({"parsed":False})
    while True:
        thread.start_new_thread(process,(pages,cur.next()))
        #process(pages2,cur.next())
        time.sleep(1)

