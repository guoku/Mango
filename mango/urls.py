__author__ = 'stxiong'
from django.conf.urls.defaults import *

urlpatterns = patterns('',
    (r'^entity/create/$', 'api.views.create_entity'),
    (r'^entity/(?P<entity_id>\d+)/$', 'api.views.read_entity'),
    (r'^entity/(?P<entity_id>\d+)/item/(?P<item_id>\w+)/unbind/$', 'api.views.unbind_entity_item'),
    (r'^entity/$', 'api.views.read_entities'),
    (r'^taobao/item/check/(?P<taobao_id>\w+)/exist/$', 'api.views.check_taobao_item_exist'),
)


