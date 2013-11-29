#encoding=utf-8
import urllib2
from bs4 import BeautifulSoup
from urlparse import parse_qs
from urlparse import urlparse
import re

def fetch(shopurl):
    html = urllib2.urlopen(shopurl).read()
    soup = BeautifulSoup(html)
    titletag = soup.findChild('p',{'class':'box'}).text
    title = titletag[3:]
    pictag = soup.findChild('td',{'class':'pic'})
    piclink = pictag.img['src']
    nick = title
    if shopurl.find('taobao')>=0:
        ownertag = soup.findChild('td',{"valign":"top"})
        owner = ownertag.next
        nick = owner.strip()[3:]

    wwimg = soup.findChild('img',{'alt':'ww'})
    pgtor = wwimg.parentGenerator()
    sidtag = pgtor.next()['href']
    uparser = urlparse(sidtag)
    shopid = parse_qs(uparser.query)['shopId'][0]
    shopid = int(shopid)
    scoretag = pgtor.next().text
    scorearray = re.findall(ur'\d\.\d分',scoretag)
    item_score = float(scorearray[0][:-1])
    service_score = float(scorearray[1][:-1])
    delivery_score = float(scorearray[2][:-1])

    print title 
    print nick
    print piclink
    print shopid
    print item_score
    print service_score
    print delivery_score


if __name__=="__main__":
    fetch("http://shop71839143.m.taobao.com")

