# coding=utf8
from models import Entity as EntityModel
from item import Item
import datetime 

class Entity(object):
    
    def __init__(self, entity_id):
        self.__entity_id = int(entity_id)
    
    def get_entity_id(self):
        return self.__entity_id
    
    def add_taobao_item(self, taobao_item_info):
        _taobao_item_obj = Item.create_taobao_item( 
            entity_id = self.__entity_id,
            taobao_id = taobao_item_info["taobao_id"],
            cid = taobao_item_info["cid"],
            title = taobao_item_info["title"],
            shop_nick = taobao_item_info["shop_nick"], 
            price = taobao_item_info["price"], 
            soldout = taobao_item_info["soldout"], 
        )
        return _taobao_item_obj.get_item_id()

    def del_taobao_item(self, item_id):
        _item_obj = Item(item_id)
        if _item_obj.get_entity_id() == self.__entity_id:
            _item_obj.bind_entity(-1)
    
    @classmethod
    def create_by_taobao_item(cls, brand, title, intro, taobao_item_info):
        
        if brand != None: 
            brand = brand.strip()
        if title != None:
            title = title.strip()
        if intro != None:
            intro = intro.strip()
        
        _entity_obj = EntityModel.objects.create( 
            brand = brand,
            title = title,
            intro = intro,
        )
        
        _inst = cls(_entity_obj.id)
        _inst.__entity_obj = _entity_obj
        
        try:
            _taobao_item_id = _inst.add_taobao_item(taobao_item_info)
        except Exception, e:
            _entity_obj.delete()
            raise e

        return _inst

    def __ensure_entity_obj(self):
        if not hasattr(self, '__entity_obj'):
            self.__entity_obj = EntityModel.objects.get(pk = self.__entity_id) 
    
    def read(self):
        self.__ensure_entity_obj()
        _context = {}
        _context["entity_id"] = self.__entity_obj.id
        _context["brand"] = self.__entity_obj.brand 
        _context["title"] = self.__entity_obj.title
        _context["intro"] = self.__entity_obj.intro
        _context["item_id_list"] = Item.get_item_id_list_by_entity_id(self.__entity_id) 
        return _context    
    
    def update(self, brand = None, title = None, intro = None):
        self.__ensure_entity_obj()
        if brand != None:
            self.__entity_obj.brand = brand
        if title != None:
            self.__entity_obj.title = title
        if intro != None:
            self.__entity_obj.intro = intro
        self.__entity_obj.save()
