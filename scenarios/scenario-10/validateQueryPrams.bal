import ballerina/http;

public function validateRequest (http:Caller outboundEp, http:Request req) {
    map<string> params = req.getQueryParams();
    string? paramStatus = params["status"];
    if (paramStatus is ()) {
        http:Response res = new;
        res.statusCode = 400;
        json message = {"error": "Missing a required parameter"};
        res.setPayload(message);
        var status = outboundEp->respond(res);
    } else if (paramStatus != "available" && paramStatus != "pending" && paramStatus != "sold" && paramStatus != "soon") {
        http:Response res = new;
        res.statusCode = 422;
        json message = {"error": "Invalid status parameter"};
        res.setPayload(message);
        var status = outboundEp->respond(res);
    }
}