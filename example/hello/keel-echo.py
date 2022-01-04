from flask import Flask, jsonify, request
app = Flask(__name__)


@app.route('/v1/identify', methods=['GET'])
def identify():
    print("identify", flush=True)
    return jsonify(
        {
            "res": {
                "ret": 0,
                "msg": "ok"
            },
            "plugin_id": "keel-echo",
            "verison": "v0.3.0",
            "tkeel_version": "v0.3.0",
            "entries": [
                {
                    "id": "echo-test-users",
                    "name": "echo-users",
                    "icon": "",
                    "path": "/users",
                    "entry": "https://tkeel-console-plugin-users.pek3b.qingstor.com/index.html"
                },
                {
                    "id": "echo-test",
                    "name": "echo-test",
                    "icon": "",
                    "children": [
                        {
                            "id": "echo-test-plugins",
                            "name": "echo-test-plugins",
                            "icon": "",
                            "path": "/plugins",
                            "entry": "https://tkeel-console-plugin-plugins.pek3b.qingstor.com/index.html"
                        }
                    ]
                }
            ]
        }
    )


@app.route('/v1/status', methods=['GET'])
def status():
    print("status", flush=True)
    return jsonify({"res": {"ret": 0, "msg": "ok"}, "status": 3})


@app.route('/v1/tenant/bind', methods=['POST'])
def tenant_bind():
    print("tenant/bind", flush=True)
    return jsonify({"res": {"ret": 0, "msg": "ok"}})


@app.route('/v1/tenant/unbind', methods=['POST'])
def tenant_unbind():
    print("tenant/unbind", flush=True)
    return jsonify({"res": {"ret": 0, "msg": "ok"}})


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
