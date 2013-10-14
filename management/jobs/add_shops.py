#coding=utf-8
import MySQLdb
import httplib

mysql_host = "localhost"
mysql_user = "root"
mysql_password = "123456"
mysql_database = "guoku_07_24_slim"

conn = MySQLdb.Connection(mysql_host, mysql_user, mysql_password, mysql_database)
cur = conn.cursor(MySQLdb.cursors.DictCursor)


cur.execute("SET names utf8")
cur.execute("SELECT nick FROM base_taobao_shop")

rows = cur.fetchall()
c = httplib.HTTPSConnection("10.0.1.103:8080")

for row in rows:
    print row['nick']
    c.request("GET", "/scheduler/api/add_shop?shop_name=%s&token=d61995660774083ccb8b533024f9b8bb" % row['nick'])
    res = c.getresponse()
    data = res.read()
    print data
