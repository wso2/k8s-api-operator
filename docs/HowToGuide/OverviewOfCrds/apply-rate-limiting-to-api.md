### Applying ratelimiting to APIs

Ratelimiting policies can be applied to the APIs created with the kubernetes operator, to throttle out requests according to the desired limit.

Ratelimiting controller obtains the policies in the below 2 methods.
* User defined policies 
* Set of default policies

The ratelimiting controller uses the policies deployed to create a policy template configmap, which gets updated when user deploy ratelimiting policies.
This policy cofigmap is mounted to kaniko job by the API controller and used to create the docker image.

#### How to deploy a ratelimiting policy

Create a ratelimiting policy using the RateLimiting kind. You can create 3 types of policies with the RateLimiting kind as shown below.
Include the limit from which you want to throttle the API requests.

* Application throttling
```
apiVersion: wso2.com/v1alpha1
kind: RateLimiting
metadata:
  name: app4
  namespace: wso2-test-ns
spec:
  type: application             
  description: Allow 4 requests per minute   # optional
  timeUnit: min
  unitTime: 1
  requestCount:
   limit: 4
```

* Subscription throttling
```
apiVersion: wso2.com/v1alpha1
kind: RateLimiting
metadata:
  name: sub6
  namespace: wso2-test-ns
spec:
  type: subscription             
  description: Allow 6 requests per minute   # optional
  timeUnit: min
  unitTime: 1
  requestCount:
   limit: 6
```
* Advance throttling 
```
apiVersion: wso2.com/v1alpha1
kind: RateLimiting
metadata:
  name: advance3
  namespace: wso2-test-ns
spec:
  type: advance             
  description: Allow 3 requests per minute   # optional
  timeUnit: min
  unitTime: 1
  requestCount:
   limit: 3
```

Implementation of the ratelimiting controller is found [here](https://github.com/wso2/k8s-api-operator/blob/master/api-operator/pkg/controller/ratelimiting/ratelimiting_controller.go)

Sample ratelimiting definitions are provided [here](../../../api-operator/deploy/sample-definitions/wso2_v1alpha1_ratelimiting_cr.yaml)
