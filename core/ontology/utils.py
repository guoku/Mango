# coding=utf-8
import hmac, hashlib, time

GUOKU_HASH_KEY = "guokushigehaowangzhan"
def cal_guoku_hash(message_origin):
    time_stamp = str(int(time.time()))
    message = message_origin.encode("utf8") + time_stamp
    sign = hmac.new(GUOKU_HASH_KEY, message, hashlib.md5).hexdigest()[0:8]
    return sign
