# coding=utf8
from models import Entity as EntityModel
from item import Item
import datetime 

class Entity(object):
    
    def __init__(self, entity_id):
        self.__entity_id = int(entity_id)
    
    def get_entity_id(self):
        return self.__entity_id
    
    @classmethod
    def create_by_taobao_item(cls, title, brand, taobao_item_info):
        
        if brand != None: 
            brand = brand.strip()
        if title != None:
            title = title.strip()
        
        _entity_obj = EntityModel.objects.create( 
            brand = brand,
            title = title 
        )
        
        try:
            _taobao_item_obj = Item.create_taobao_item( 
                entity_id = _entity_obj.id, 
                taobao_id = taobao_item_info["taobao_id"],
                category_id = taobao_item_info["category_id"],
                title = taobao_item_info["title"],
                shop_nick = taobao_item_info["shop_nick"], 
                price = taobao_item_info["price"], 
                soldout = taobao_item_info["soldout"], 
            )

        except Exception, e:
            _entity_obj.delete()
            raise e

        _inst = cls(_entity_obj.id)
        _inst.__entity_obj = _entity_obj
        return _inst

    def __ensure_entity_obj(self):
        if not hasattr(self, '__entity_obj'):
            self.__entity_obj = Entity.objects.get(pk = self.__entity_id) 
    
    def read(self):
        _context = {}
        
        _context["entity_id"] = self.__entity_obj.id
        _context["brand"] = self.__entity_obj.brand 
        _context["title"] = self.__entity_obj.title
        _context["created_time"] = self.__entity_obj.created_time
        _context["updated_time"] = self.__entity_obj.updated_time
            
        return _context    
