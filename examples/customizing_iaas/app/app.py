import json
import sys


data = json.loads(sys.stdin.read())

result = {
    'result': pow(data['a'], data['b']),
}


ret = json.dumps(result)
sys.stdout.write(ret)
