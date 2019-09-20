# k8s-apim-operator Scenarios

## Scenario 1

> ##### This scenario demonstrates Basic API deployment

- Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics
 
##### Navigate to the scenarios/scenario-1 directory and execute the following command

- Create API <br /> 
    - ***apimcli add api -n petstore --from-file=petstore_basic.yaml***

- Update API <br /> 
    - ***apimcli update api -n petstore --from-file=petstore_basic.yaml***
    
- Get available API <br /> 
    - ***apimcli get apis***

- Get service details to invoke the API<br />
    - ***apimcli get services***
    - Sample Output:
    
    ```
    NAME          TYPE         CLUSTER-IP      EXTERNAL-IP          PORT(S)                     AGE
    petstore   LoadBalancer    10.0.3.74     104.199.77.249   9095:30453/TCP,9090:32422/TCP     1m
    ```

    - Note: Get the external IP of the service
 
- Invoking the API <br />
    - ***TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UQXhabU14TkRNeVpEZzNNVFUxWkdNME16RXpPREpoWldJNE5ETmxaRFUxT0dGa05qRmlNUSJ9.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbiIsImFwcGxpY2F0aW9uIjp7ImlkIjoyLCJuYW1lIjoiSldUX0FQUCIsInRpZXIiOiJVbmxpbWl0ZWQiLCJvd25lciI6ImFkbWluIn0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC9sb2NhbGhvc3Q6OTQ0M1wvb2F1dGgyXC90b2tlbiIsImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6Ilg5TGJ1bm9oODNLcDhLUFAxbFNfcXF5QnRjY2EiLCJleHAiOjM3MDMzOTIzNTMsImlhdCI6MTU1NTkwODcwNjk2MSwianRpIjoiMjI0MTMxYzQtM2Q2MS00MjZkLTgyNzktOWYyYzg5MWI4MmEzIn0=.b_0E0ohoWpmX5C-M1fSYTkT9X4FN--_n7-bEdhC3YoEEk6v8So6gVsTe3gxC0VjdkwVyNPSFX6FFvJavsUvzTkq528mserS3ch-TFLYiquuzeaKAPrnsFMh0Hop6CFMOOiYGInWKSKPgI-VOBtKb1pJLEa3HvIxT-69X9CyAkwajJVssmo0rvn95IJLoiNiqzH8r7PRRgV_iu305WAT3cymtejVWH9dhaXqENwu879EVNFF9udMRlG4l57qa2AaeyrEguAyVtibAsO0Hd-DFy5MW14S6XSkZsis8aHHYBlcBhpy2RqcP51xRog12zOb-WcROy6uvhuCsv-hje_41WQ==***
   
    - ***curl -X GET "https://\<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k***

- Delete API <br /> 
    - ***apimcli delete api petstore***