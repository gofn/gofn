#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json

obj = {"itens": ["a", "b", "c"], "boolean": True, "interge": 123456}

ret = json.dumps(obj, ensure_ascii=False)
print(ret)
# print('{"itens": ["a", "b", "c"], "boolean": true, "interge": 123456}')
