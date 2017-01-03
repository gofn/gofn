#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import json
from random import randint
import time

time.sleep(5)  # Stop for a while to simulate some processing

obj = {"items": ["a", "b", "c"], "boolean": True, "integer": 123456,
       "random": randint(0, 9999)}

ret = json.dumps(obj, ensure_ascii=True)
print(ret)
