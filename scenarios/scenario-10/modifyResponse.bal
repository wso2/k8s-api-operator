import ballerina/http;

public function validateResponse (http:Caller outboundEp, http:Response res) {
    var payload = res.getJsonPayload();
    if (payload is json) {
        json resJson = {
            "pets": payload,
            "length": payload.length()
        };
        res.setJsonPayload(untaint resJson);
    }
}