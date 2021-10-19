from flask import Flask,jsonify,request

app = Flask(__name__)

@app.route('/v1/identify',methods=['GET'])
def identify():
    print("identify", flush=True)
    return jsonify({"ret": 0,"msg": "ok","plugin_id": "demo-echo","version": "0.0.1","main_plugins": [{"id": "keel","version": "1.0","endpoints": [{"addons_point": "externalPreRouteCheck","endpoint": "echo"}]}]})

@app.route('/v1/status',methods=['GET'])
def status():
    print("status", flush=True)
    return jsonify({"ret":0,"msg":"ok","status":"ACTIVE"})

@app.route('/echo',methods=['GET','POST','DELETE','OPTION','PUT'])
def echo():
    print("echo", flush=True)
    if 'x-keel-check' in request.headers:
        print("keel registered check")
        check_header=request.headers.get('x-keel-check')
        print("header is "+check_header,flush=True)
        if check_header == 'True':
            return jsonify({"msg":"ok","ret":0})
        else:
            return jsonify({"msg":"faild","ret":-1})
    print(request.get_data(), flush=True)
    return jsonify({"msg":"ok","ret":0})

if __name__ == '__main__':
    app.run(port=8080)
