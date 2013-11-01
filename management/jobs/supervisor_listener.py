#coding=utf-8
from pymongo import Connection
import datetime
import xmlrpclib
from threading import Timer
import time
def detect():

	con = Connection()

	db = con.test

	status = db.status

	item = status.find_one()

	timestamp = item['timestamp']

	now = datetime.datetime.now()

	timedelta = now - timestamp

	deltasec = timedelta.seconds

	if deltasec > 1800:
		#超过半个小时爬虫没有发送数据就让supervisor重启爬虫
		s = xmlrpclib.Server("http://guoku:123456@10.0.1.23:9001/RPC2")
		#statuinfo = s.supervisor.getProcessInfo("node_crawler")
		#state = statuinfo['statename']
		#if state=='RUNNING':
		try:
			s.supervisor.stopProcessGroup("node_crawler")
			print("stop process group node_crawler")
		finally:

			s.supervisor.startProcessGroup("node_crawler")
			print("start process group node_crawler")
		
			#s.supervisor.startProcessGroup("node_crawler")
			#print("start process group node_crawler")


def timer():
	t = Timer(1800,detect())
	t.start()

if __name__=='__main__':
	timer()
