from models import Entity as EntityModel

class Entity(object):
    
    def __init__(self, entity_id):
        self.__entity_id = int(entity_id)
    
#    @classmethod
#    def cal_entity_hash(cls, entity_hash_string):
#        while True:
#            entity_hash = utils.cal_guoku_hash(entity_hash_string)
#            try:
#                Entity.objects.get(entity_hash = entity_hash)
#            except:
#                break
#        return entity_hash
    
