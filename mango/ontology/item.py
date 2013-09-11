# coding=utf8
from models import Item as ItemDocument
from models import TaobaoItem as TaobaoItemDocument
import datetime

class Item(object):
    
    def __init__(self, item_id):
        self.__item_id = item_id 
    
    def get_item_id(self):
        return self.__item_id
    
    def __ensure_item_obj(self):
        if not hasattr(self, '__item_obj'):
            self.__item_obj = ItemDocument.objects.filter(id = self.__item_id).first()
    
    def get_entity_id(self):
        self.__ensure_item_obj()
        return self.__item_obj.entity_id
    
    @staticmethod
    def get_item_id_list_by_entity_id(entity_id):
        _list = []
        for _item in ItemDocument.objects.filter(entity_id = entity_id):
            _list.append(str(_item.id))
        return _list

    @staticmethod
    def get_entity_id_by_taobao_id(taobao_id):
        _taobao_item_obj = TaobaoItemDocument.objects.filter(taobao_id = taobao_id).first()
        if _taobao_item_obj != None:
            return _taobao_item_obj.entity_id
        return None

    @classmethod
    def create_taobao_item(cls, entity_id, taobao_id, cid, title, shop_nick, price, soldout): 
        _taobao_id = taobao_id.strip()
        _title = title.strip()
        _shop_nick = shop_nick.strip()
        
        _item_obj = TaobaoItemDocument( 
            entity_id = entity_id,
            source = 'taobao',
            taobao_id = _taobao_id,
            cid = cid,
            title = _title,
            shop_nick = _shop_nick,
            price = price,
            soldout = soldout,
            created_time = datetime.datetime.now(),
            updated_time = datetime.datetime.now() 
        )
        _item_obj.save()

        _inst = cls(_item_obj.id)
        _inst.__item_id = _item_obj.id
        _inst.__item_obj = _item_obj
        return _inst

    def bind_entity(self, entity_id):
        self.__ensure_item_obj()
        self.__item_obj.entity_id = entity_id 
        self.__item_obj.save()
    
    
    def __load_taobao_item(self):
        _context = {}
        _context["item_id"] = str(self.__item_obj.id)
        _context["entity_id"] = self.__item_obj.entity_id
        _context["source"] = self.__item_obj.source
        _context["taobao_id"] = self.__item_obj.taobao_id
        _context["cid"] = self.__item_obj.cid
        _context["title"] = self.__item_obj.title
        _context["shop_nick"] = self.__item_obj.shop_nick
        _context["price"] = float(self.__item_obj.price)
        _context["soldout"] = self.__item_obj.soldout
        return _context


    def read(self):
        self.__ensure_item_obj()
        if self.__item_obj.source == 'taobao':
            _context = self.__load_taobao_item()
        return _context    
