#encoding=utf-8

from bs4 import BeautifulSoup
import re
from urlparse import parse_qs
from urlparse import urlparse
import pymongo

states = set([u"北京",u"上海",u"天津",u"重庆",u"广东",u"江苏",u"浙江",u"山东",u"河北",u"山西",u"辽宁"
    ,u"吉林",u"河南",u"安徽",u"福建",u"江西",u"黑龙江",u"湖南",u"湖北",u"海南",u"四川",u"贵州",u"云南",
    u"陕西",u"甘肃",u"青海",u"台湾",u"西藏",u"内蒙古",u"广西",u"宁夏",u"新疆",u"香港",u"澳门"])

def fetch(html):
    soup = BeautifulSoup(html)
    
    desc = soup.title.string[0:-9]

    cattag = soup.p
    atag = cattag.findChildren('a')
    cidtag = atag[2]
    cidurl = cidtag.attrs['href']
    cid = int(re.findall(r'\d+$',cidurl)[0])

    firstcat = atag[1].text
    secondcat = cidtag.text

    category = [firstcat,secondcat]

    details = soup.find_all('div', { 'class' : 'detail' })
        
    imgurls = []      
    imgtag = details[0]
    src = imgtag.img['src']
    jpgindex = src.index('.jpg')
    if jpgindex>0:
        src = src[0:jpgindex+4]
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
    judge = re.findall(ur'格：',detail.p.text)
    hasprom = True #默认都有促销
    if len(judge)>0:
        hasprom = False
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
        price =re.findall(r'\d+\.\d+',pricetag)
        if len(price)>0:
            price = float(price[0])

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

    reviewtag = soup.findChildren('td','link_btn fix_btn')[1]
    reviews = int(re.findall(r'\d+',reviewtag.a.span.text)[0])
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
            "nick":nick
            } 
    print city
    print state
    return result


def fetchdetail(html):
    soup = BeautifulSoup(html)

    boxs = soup.find_all('div',{'class':'box'})

    reviewtag = boxs[2].p.a.strong.text

    reviewregx = re.findall('\d+',reviewtag)
    reviews = 0
    if len(reviewregx)>0:
        reviews = int(reviewregx[0])

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

def loaditem():
    conn = pymongo.Connection()
    db = conn['zerg']
    pages = db['pages'] 
    item = pages.find_one({'parsed':False})
    return item

def process():
    item = loaditem()
    print item.keys()
    fontdata = fetch(item['fontpage'])
    detaildata = fetchdetail(item['detailpage'])
    if detaildata['reviews']>0:
        fontdata['reviews']=detaildata['reviews']
    
#fontdata =fetch("16356882686")
#detaildata =  fetchdetail("16356882686")
#if detaildata["reviews"]>0:
#    fontdata["reviews"]=detaildata["reviews"]
#    print fontdata['reviews']

process()

