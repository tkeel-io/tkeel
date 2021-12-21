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
                    "path": "/users",
                    "entry": "https://tkeel-console-plugin-users.pek3b.qingstor.com/index.html",
                    "menu": [
                        "echo",
                        "test-user"
                    ]
                },
                {
                    "id": "echo-test-plugins",
                    "name": "echo-plugins",
                    "path": "/plugins",
                    "entry": "https://tkeel-console-plugin-plugins.pek3b.qingstor.com/index.html",
                    "menu": [
                        "echo",
                        "test-plugins"
                    ]
                }
            ]
        }
    )


@app.route('/v1/status', methods=['GET'])
def status():
    print("status", flush=True)
    return jsonify({"res": {"ret": 0, "msg": "ok"}, "status": 3})


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
