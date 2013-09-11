# coding=utf8
from django.db import models
from mongoengine import *

class Entity(models.Model):
    brand = models.CharField(max_length = 256, null = True)
    title = models.CharField(max_length = 256, null = True)
    created_time = models.DateTimeField(auto_now_add = True, db_index = True)
    updated_time = models.DateTimeField(auto_now = True, db_index = True)
    
class Item(Document):
    entity_id = IntField(required = True) 
    source = StringField(required = True)
    created_time = DateTimeField(required = True)
    updated_time = DateTimeField(required = True)
    meta = {
        "indexes" : [ 
            "entity_id" 
        ],
        "allow_inheritance" : True
    }

class TaobaoItem(Item):
    taobao_id = StringField(required = True, unique = True)
    cid = IntField(required = True) 
    title = StringField(required = True)
    shop_nick = StringField(required = True)
    price = DecimalField(required = True)
    soldout = BooleanField(required = True) 

    meta = {
        'indexes' : [ 
            'taobao_id',
            'cid',
            'shop_nick',
            'price',
            'soldout'
        ],
    }
    
