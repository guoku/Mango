# coding=utf8

from django.core.urlresolvers import reverse
from django.http import HttpResponseRedirect, HttpResponse, Http404
from django.shortcuts import render_to_response
from django.template import RequestContext

from ontology.entity import Entity
from ontology.item import Item 
from utils.http import JSONResponse, SuccessJsonResponse, ErrorJsonResponse 

def create_entity(request):
    try:
        if request.method == 'POST':
            _entity = Entity.create_by_taobao_item(
                brand = request.POST.get('brand', None),
                title = request.POST.get('title', None),
                intro = request.POST.get('intro', None),
                taobao_item_info = { 
                    'taobao_id' : request.POST.get('taobao_id', None),
                    'cid' : request.POST.get('cid', None),
                    'title' : request.POST.get('taobao_title', None),
                    'shop_nick' : request.POST.get('taobao_shop_nick', None),
                    'price' : float(request.POST.get('taobao_price', None)),
                    'soldout' : bool(int(request.POST.get('taobao_soldout', '0'))),
                },
            )
            _data = { "entity_id" : _entity.get_entity_id() }
            return SuccessJsonResponse(_data)
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))
        
def update_entity(request, entity_id):
    try:
        if request.method == 'POST':
            Entity(entity_id).update(
                brand = request.POST.get('brand', None),
                title = request.POST.get('title', None),
                intro = request.POST.get('intro', None),
            )
            return SuccessJsonResponse({})
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))



def add_taobao_item_for_entity(request, entity_id):
    try:
        if request.method == 'POST':
            _entity = Entity(entity_id)
            _item_id = _entity.add_taobao_item(
                taobao_item_info = { 
                    'taobao_id' : request.POST.get('taobao_id', None),
                    'cid' : request.POST.get('cid', None),
                    'title' : request.POST.get('taobao_title', None),
                    'shop_nick' : request.POST.get('taobao_shop_nick', None),
                    'price' : float(request.POST.get('taobao_price', None)),
                    'soldout' : bool(int(request.POST.get('taobao_soldout', '0'))),
                }
            )
            _data = { "item_id" : _item_id }
            return SuccessJsonResponse(_data)
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))
        
def check_taobao_item_exist(request, taobao_id):
    try:
        _entity_id = Item.get_entity_id_by_taobao_id(taobao_id)
        if _entity_id != None:
            _data = { 
                "exist" : 1,
                "entity_id" : _entity_id,
            }
        else:
            _data = { "exist" : 0 }
        return SuccessJsonResponse(_data)
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))
    
    
def read_entity(request, entity_id):
    try:
        if request.method == 'GET':
            _entity = Entity(entity_id)
            return SuccessJsonResponse({ "context" : _entity.read() })
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))

def read_entities(request):
    try:
        if request.method == 'GET':
            _entity_id_list = request.GET.getlist("eid")
            _rslt = {}
            for _entity_id in _entity_id_list:
                try:
                    _entity = Entity(_entity_id)
                    _entity_context = _entity.read()
                    _rslt[_entity_id] = {
                        'context' : _entity_context,
                        'status' : '0'
                    }
                except Exception, e:
                    _rslt[_entity_id] = {
                        'msg' : str(e),
                        'status' : '1'
                    }
                    
            return SuccessJsonResponse(_rslt)
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))

def read_items(request):
    try:
        if request.method == 'GET':
            _item_id_list = request.GET.getlist("iid")
            _rslt = {}
            for _item_id in _item_id_list:
                try:
                    _item = Item(_item_id)
                    _item_context = _item.read()
                    _rslt[_item_id] = {
                        'context' : _item_context,
                        'status' : '0'
                    }
                except Exception, e:
                    _rslt[_item_id] = {
                        'msg' : str(e),
                        'status' : '1'
                    }
                    
            return SuccessJsonResponse(_rslt)
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))

def unbind_entity_item(request, entity_id, item_id):
    try:
        _entity = Entity(entity_id)
        _entity.del_taobao_item(item_id)
        return SuccessJsonResponse({})
    except Exception, e:
        return ErrorJsonResponse(emsg = str(e))
     
