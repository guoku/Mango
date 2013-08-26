from models import TaobaoItem as TaobaoItemDocument
import datetime

class Item(object):
    
    def __init__(self, item_id):
        self.__item_id = item_id 
    
    def get_item_id(self):
        return self.__item_id
    
    @classmethod
    def create_taobao_item(cls, entity_id, taobao_id, shop_nick ): 

        _item_doc = TaobaoItemDocument( 
            entity_id = entity_id,
            taobao_id = taobao_id,
            shop_nick = shop_nick,
            created_time = datetime.datetime.now(),
            updated_time = datetime.datetime.now() 
        )
        
        _inst = cls(_item_doc.id)
        _inst.__item_doc = _item_doc
        return _inst
