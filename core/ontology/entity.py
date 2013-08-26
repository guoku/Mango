from models import Entity as EntityModel
from item import Item
import datetime 
import utils

class Entity(object):
    
    def __init__(self, entity_id):
        self.__entity_id = int(entity_id)
    
    @classmethod
    def cal_entity_hash(cls, entity_hash_string):
        while True:
            entity_hash = utils.cal_guoku_hash(entity_hash_string)
            try:
                Entity.objects.get(entity_hash = entity_hash)
            except:
                break
        return entity_hash
    
    def get_entity_id(self):
        return self.__entity_id
    
    @classmethod
    def create_by_taobao_item(cls, title, brand, taobao_item_info):
        
        entity_hash = cls.cal_entity_hash(taobao_item_info["taobao_id"])
    
        brand = brand.strip() 
        title = title.strip() 
        _entity_obj = EntityModel.objects.create( 
            entity_hash = entity_hash, 
            brand = brand,
            title = title )

        
        try:
            _taobao_item_obj = Item.create_taobao_item( 
                entity_id = _entity_obj.id, 
                taobao_id = taobao_item_info["taobao_id"],
                shop_nick = taobao_item_info["shop_nick"] 
            )


        except Exception, e:
            _entity_obj.delete()
            raise e

        _inst = cls(_entity_obj.id)
        _inst.__entity_obj = _entity_obj
        return _inst
