__author__ = 'stxiong'
from django.conf.urls.defaults import *

urlpatterns = patterns('',
    (r'^entity/create/$', 'api.views.create_entity'),
)


