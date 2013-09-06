# coding=utf8

from django.core.urlresolvers import reverse
from django.http import HttpResponseRedirect, HttpResponse, Http404
from django.shortcuts import render_to_response
from django.template import RequestContext

from ontology.entity import Entity
from ontology.item import Item 
from utils.http import JSONResponse, SuccessJsonResponse, ErrorJsonResponse 

def create_entity(request):
    if request.method == 'POST':
        _entity = Entity.create_by_taobao_item(
            brand = request.POST.get('brand', None),
            title = request.POST.get('title', None),
            taobao_item_info = { 
                'taobao_id' : request.POST.get('taobao_id', None),
                'category_id' : request.POST.get('taobao_category_id', None),
                'title' : request.POST.get('taobao_title', None),
                'shop_nick' : request.POST.get('taobao_shop_nick', None),
                'price' : float(request.POST.get('taobao_price', None)),
                'soldout' : bool(int(request.POST.get('taobao_soldout', '0'))),
            },
        )
        _data = { "entity_id" : _entity.get_entity_id() }
        return SuccessJsonResponse(_data)

def check_taobao_item_exist(request, taobao_id):
    _entity_id = Item.get_entity_id_by_taobao_id(taobao_id)
    if _entity_id != None:
        _data = { 
            "exist" : 1,
            "entity_id" : _entity_id,
        }
    else:
        _data = { "exist" : 0 }
    return SuccessJsonResponse(_data)
    
    
