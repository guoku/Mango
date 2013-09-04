# coding=utf8
from django.http import HttpResponse
from django.utils import simplejson as json


class JSONResponse(HttpResponse):
    def __init__(self, result = {}, **kwargs):
        _content = json.dumps(result, indent = 2, ensure_ascii = False)
        kwargs['content_type'] = 'application/json; charset=utf8'
        super(JSONResponse, self).__init__(_content, **kwargs)


class SuccessJsonResponse(JSONResponse):

    def __init__(self, data = []):
        _res = {}
        _res['res_code'] = 0
        _res['res_msg'] = 'success'
        _res['data'] = data
        super(SuccessJsonResponse, self).__init__(_res)


class ErrorJsonResponse(JSONResponse):
    def __init__(self, data = [], emsg = None):
        _res = {}
        _res['res_code'] = 1 
        _res['res_msg'] = 'error' 
        _res['data'] = data
        super(ErrorJsonResponse, self).__init__(_res)

class JSONResponseParser(object):
    def __init__(self, response):
        _content = json.loads(response)
        if _content['res_code'] == 0:
            self.__status = 0
            self.__data = _content['data'] 
        else:
            self.__status = 1

    def success(self):
        if self.__status == 0:
            return True
        return False

    def read(self):
        return self.__data
            
            
        
    
