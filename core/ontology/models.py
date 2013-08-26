from django.db import models
from mongoengine import *

class Entity(models.Model):
    entity_hash = models.CharField(max_length = 32, unique = True, db_index = True)
    brand = models.CharField(max_length = 256, null = True)
    title = models.CharField(max_length = 256, null = True)
    created_time = models.DateTimeField(auto_now_add = True, db_index = True)
    updated_time = models.DateTimeField(auto_now = True, db_index = True)
    
class Item(Document):
    item_id = IntField(required = True) 
    entity_id = IntField(required = True) 
    created_time = DateTimeField(required = True)
    updated_time = DateTimeField(required = True)
    meta = {
        "indexes" : [ 
            "item_id", 
            "entity_id" 
        ],
        "allow_inheritance" : True
    }

class TaobaoItem(Item):
    taobao_id = StringField(required = True)
    shop_nick = StringField(required = True)
    meta = {
        "indexes" : [ 
            "taobao_id",
            "shop_nick" 
        ],
    }
    
