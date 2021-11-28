from flask import Flask, jsonify, request
app = Flask(__name__)


@app.route('/v1/identify', methods=['GET'])
def identify():
    print("identify", flush=True)
    return jsonify({"res": {"ret": 0, "msg": "ok"}, "plugin_id": "keel-echo", "verison": "v0.2.0", "tkeel_version": "v0.2.0"})


@app.route('/v1/status', methods=['GET'])
def status():
    print("status", flush=True)
    return jsonify({"res": {"ret": 0, "msg": "ok"}, "status": 2})


@app.route('/echo', methods=['GET', 'POST', 'DELETE', 'OPTION', 'PUT'])
def echo():
    queryStr = str(request.query_string, encoding="utf-8")
    header = request.headers.__str__()
    print({"query_string": queryStr}, flush=True)
    data = str(request.get_data(), encoding="utf-8")

    print({"data": data}, flush=True)
    return jsonify({"query_string": queryStr, "data": data, "header": header})


if __name__ == '__main__':
    app.run(port=8080)
