#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json
from random import randint
import time 

time.sleep(5) # Stop for a while to simulate some processing

obj = {"itens": ["a", "b", "c"], "boolean": True, "interge": 123456, "random": randint(0,9999)}

ret = json.dumps(obj, ensure_ascii=False)
print(ret)
# print('{"itens": ["a", "b", "c"], "boolean": true, "interge": 123456}')
