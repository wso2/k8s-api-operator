# k8s-apim-operator Scenarios

## Scenario 4

> ##### This scenario demonstrates ratelimiting (throttling), OAuth2 flow, Analytics, Security kind, API addition, update, show and delete

- Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics
 
##### Navigate to the k8s-apim-operator/apim-operators/scenarios/scenario-4 directory and execute the following command

- Create namespaces <br /> 
    - ***kubectl apply -f namespace.yaml***

- Create few ratelimiting instances in the abobe namespaces. <br /> 
    - ***kubectl apply -f ratelimiting_instances.yaml***
    - Note: Samples are given for all 3 types of ratelimiting kind: advance, subscription, application. Using advance3 and advance5 - two advance ratelimiting instances is sufficient to demonstrate this scenario. To work with application and subscription ratelimiting types, user need to have created those policies in the APIM server

- Create OAuth2 type instances of security kind and secrets with credentials for the APIM server. <br />
    - ***kubectl apply -f oauth_security.yaml***
    - Note: User must deploy two secrets with the certificates. The following command can be used to a create secret from the certificate pem file.
    - ***kubectl create secret generic wso2am260-secret --from-file=wso2am260.pem -n=wso2-test-ns1***
    - ***kubectl create secret generic wso2am260-secret --from-file=wso2am260.pem -n=wso2-test-ns2***

- Do hostmapping for APIM and analytics servers.  <br />
    - ***kubectl apply -f host_mapping.yaml***

- Deploy two APIs in the created namespaces.  <br />
    - ***kubectl add api myapi1 --from-file=swagger_ns1_oauth.yaml -n=wso2-test-ns1***
    - ***kubectl add api myapi2 --from-file=swagger_ns2_oauth.yaml -n=wso2-test-ns1***

- Update the two APIs.  <br />
    - ***kubectl update api myapi1 --from-file=swagger_ns1_oauth.yaml -n=wso2-test-ns1 --replicas=2***
    - ***kubectl update api myapi2 --from-file=swagger_ns2_oauth.yaml -n=wso2-test-ns1 --replicas=2***

- View API and the used swagger.  <br />
    - ***kubectl show api myapi1 -n=wso2-test-ns1***
    - ***kubectl show api myapi2 -n=wso2-test-ns1***

- Delete the two APIs and all created resources related to the particular API.  <br />
    - ***kubectl delete api myapi1 -n=wso2-test-ns1***
    - ***kubectl delete api myapi2 -n=wso2-test-ns1***