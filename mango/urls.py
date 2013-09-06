__author__ = 'stxiong'
from django.conf.urls.defaults import *

urlpatterns = patterns('',
    (r'^entity/create/$', 'api.views.create_entity'),
    (r'^taobao/item/check/(?P<taobao_id>\w+)/exist/$', 'api.views.check_taobao_item_exist'),
)


