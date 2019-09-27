# k8s-apim-operator Scenarios

## Scenario 7

> ##### This scenario demonstrates ratelimiting(throttling) scenario

- Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics
 
##### Navigate to the scenarios/scenario-6 directory and execute the following command

- Deploy targetendpoint kind in the cluster <br /> 
    - ***apimcli apply -f privateJet-endpoint.yaml***
    
- Create API <br /> 
    - ***apimcli add api -n petstore-privatejet --from-file=swagger-privateJet-dep.yaml***

- Get target endpoint
    - ***apimcli get targetendpoints***

- Get service details to invoke the API<br />
    - ***apimcli get services***
    - Note: Get the external IP of the service
 
- Invoking the API <br />
    - ***TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UQXhabU14TkRNeVpEZzNNVFUxWkdNME16RXpPREpoWldJNE5ETmxaRFUxT0dGa05qRmlNUT09In0=.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJKV1QtQWxwaGEiLCJpZCI6NH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTo5NDQzXC9vYXV0aDJcL3Rva2VuIiwidGllckluZm8iOnt9LCJrZXl0eXBlIjoiUFJPRFVDVElPTiIsInN1YnNjcmliZWRBUElzIjpbXSwiY29uc3VtZXJLZXkiOiJFaFY5QzNfcWhFbk1jN3J3ajJnc2VqeWVfdW9hIiwiZXhwIjozNzE2NDY3ODA0LCJpYXQiOjE1Njg5ODQxNTcsImp0aSI6IjgwNGMwNDk0LTAwNWYtNDE3MS1hMDY1LTc5OGViZTBlYzM1YiJ9.gzpBiMaa6UVKA27UJpp1zjhxKiAWY_Zq_pMFB5n4hEJ3ydrekdXk5kvukinBFQXeLHUEdREBFHhqWVxFosuoM25UVbo55PJn_XtVZ42j-AwO3cTdIC73eC-OgwdCkoXJdf4wSRXpJNZj66iItxyhGTvxVwDfom43TCreLJfqyEbpJ6NixjWz4seU0YUJXOC8se5EAAhS2gtbESdQNugx0E7XobjYjd-sG7qj_Mhvbq6N0SZ79Eko-QWXqb3BF98vslBERKWh7h6b-JDjO5Lr-pEizWEBGdz1mNSQz2qmxZYrFsigPNPkYG-IrwpaITyrVLGMLQQGoT4DHpLsVT1bqw==***
   
    - ***curl -X GET "https://\<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k***

- Delete API <br /> 
    - ***apimcli delete api petstore-privatejet***
