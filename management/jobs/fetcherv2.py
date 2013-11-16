#encoding=utf-8

import urllib2
from bs4 import BeautifulSoup
import re
from urlparse import parse_qs
from urlparse import urlparse
class crawler:
    desc= ''
    cid = ''
    category = []
    imgurls = []
    promprice = ''
    price = ''
    salecount = ''
    location = ''
    reviesws = ''
    nick = ''

    itemid = ''

    def __init__(self,itemid):
        self.itemid = itemid

    def fetch(self):

        response = urllib2.urlopen('http://a.m.taobao.com/i'+self.itemid+'.htm')
        html = response.read()

        soup = BeautifulSoup(html)

        self.desc = soup.title.string[0:-9]

        cattag = soup.p
        atag = cattag.findChildren('a')
        cidtag = atag[2]
        cidurl = cidtag.attrs['href']
        self.cid = re.findall(r'\d+$',cidurl)[0]

        firstcat = atag[1].text
        secondcat = cidtag.text

        self.category = [firstcat,secondcat]

        tables = soup.find_all('table')
        imgtable = tables[1]
        imgtags = imgtable.findChildren('img')
        for tag in imgtags:
            imgurl = tag['src']
            imgurl = imgurl.replace('_70x70.jpg','')
            self.imgurls.append(imgurl)

        detail = soup.find_all('div',class_ = 'detail')[1]

        prom = detail.p.strong

        if prom != None:
            self.promprice = prom.text

        p = detail.findChildren('p')
        pricetag = p[1].text

        if pricetag!=None:
            price =re.findall(r'\d+\.\d+',pricetag)
            if len(price)>0:
                self.price = price[0]

        counttag = p[3].text
        tmp = re.findall(r'\d+',counttag)
        if len(tmp)>0:
            self.salecount = tmp[0]

        loctag = p[4].text
        if loctag.find(u'地')>0:
            self.location = loctag.split(u'：')[1]

        reviewtag = soup.findChildren('td','link_btn fix_btn')[1]
        self.reviesws = re.findall(r'\d+',reviewtag.a.span.text)[0]
        nametag = soup.select('body div.bd div.box div.detail p a img')[0]
        nameurl = nametag['src']
        o = urlparse(nameurl)
        nick = parse_qs(o.query)['nick'][0]
        self.nick= nick
 

if __name__=='__main__':
    cc = crawler("19864856561")
    cc.fetch()
    print cc.desc
    print cc.cid 
    print cc.nick
    print cc.salecount
    print cc.promprice
    print cc.price
    print cc.reviesws
    print cc.location
    for i in cc.category:
        print i

    for i in cc.imgurls:
        print i
