# Create your views here.

from django.core.urlresolvers import reverse
from django.http import HttpResponseRedirect, HttpResponse, Http404
from django.shortcuts import render_to_response
from django.template import RequestContext

from ontology.entity import Entity

def create_entity(request):
    if request.method == 'POST':
        _brand = request.POST.get("brand", None)
        _title = request.POST.get("title", None)
        _taobao_id = request.POST.get("taobao_id", None)
        _taobao_shop_nick = request.POST.get("taobao_shop_nick", None)
        
        _entity = Entity.create_by_taobao_item(
            title = _title,
            brand = _brand,
            taobao_item_info = { 
                "taobao_id" : _taobao_id,
                "shop_nick" : _taobao_shop_nick
            }
        )
        return HttpResponse(_entity.get_entity_id())
