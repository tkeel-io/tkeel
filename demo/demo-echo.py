from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse

app = App()

@app.method(name='v1/identify')
def identify(request: InvokeMethodRequest) -> InvokeMethodResponse:
    print("identify", flush=True)
    return InvokeMethodResponse(b'{"ret": 0,"msg": "ok","plugin_id": "demo-echo","version": "0.0.1","main_plugins": [{"id": "keel","version": "1.0","endpoints": [{"addons_point": "externalPreRouteCheck","endpoint": "echo"}]}]}', "application/json")

@app.method(name='v1/status')
def status(request: InvokeMethodRequest) -> InvokeMethodResponse:
    print("status", flush=True)
    return InvokeMethodResponse(b'{"ret":0,"msg":"ok","status":"ACTIVE"}', "application/json")

@app.method(name='echo')
def echo(request: InvokeMethodRequest) -> InvokeMethodResponse:
    print("echo", flush=True)
    header_dict=request.get_metadata(True)
    if 'x-keel-check' in header_dict:
        print("keel registered check")
        check_header=header_dict['x-keel-check']
        if check_header[0] == 'True':
            return InvokeMethodResponse('{"msg":"ok","ret":0}',"application/json")
        else:
            return InvokeMethodResponse('{"msg":"faild","ret":-1}',"application/json")
    print(request.text(), flush=True)
    return InvokeMethodResponse('{"msg":"ok","ret":0}', "application/json")

app.run(50051)
