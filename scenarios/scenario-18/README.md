## Scenario 18 - Expose an API using Openshift Route

- This scenario shows how to expose an API using Openshift Route.

**Important:**
> Follow the main README and deploy the api-operator and configuration files.

**Prerequisites:**

- Openshift Cluster (v3.11 or higher version).

- Grant relevant privileges

    You can skip this section if you are deploying the API Operator via Operators in Openshift console.
      
    - Log in as Admin and create a project called wso2
        ```
        >> oc new-project wso2
        ```

    - Grant access to the privileged SCC
        ```
        >> oc adm policy add-scc-to-user <scc_name> <user_name>

        Eg: oc adm policy add-scc-to-user privileged kubeadmin
        ```

    - Add this command to enable container images with any user
        ```
        >> oc adm policy add-scc-to-group anyuid system:authenticated
        ```   
 
- Navigate to the api-operator/controller-artifacts directory and set the operatorMode to "Route" in the 
  controler_conf.yaml file.
  
  ```
  operatorMode: "Route"
  ```
- If you have already deployed the operator you have to update operatorMode to "Route" and apply the changes using
  following command.
  ```
  >> apictl apply -f api-operator/controller-artifacts/controler_conf.yaml
  ```
  
#### Deploying the artifacts

- Navigate to scenarios/scenario-18 directory and deploy the sample backend service using the following command.
    
    ```
    >> apictl apply -f hello-world-service.yaml
    
    Output:
    targetendpoint.wso2.com/hello-world-service created
    ```
- Basic swagger definition belongs to the "hello-world-service" service is available in swagger.yaml.
- Create an API which is refer to the backend service "hello-world-service" using following command.
 
    ```
    >> apictl add api -n hello-world-api --from-file=swagger.yaml
    
    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/hello-world-1-swagger created
    creating API definition
    api.wso2.com/hello-world created
    ```
   
- List down routes

    ```
    >> apictl get routes
    
    Output:
    NAME                                                HOST/PORT            PATH        SERVICES         PORT         TERMINATION      WILDCARD
    api-operator-route-hello-world-api-node-1.0.0     mgw.route.wso2.com   /node/1.0.0  hello-world-api   9090         edge             None
    ```
      
 - Invoking the API 

    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
    
    ```
    >> curl -H "Host:mgw.route.wso2.com" https://34.67.56.7/node/1.0.0/hello/node -H "Authorization:Bearer $TOKEN" -k
    
    Output:
    Hello World!
    ```

**Notes** 
- Only TLS edge support is provided. 
- Tested in Openshift v4.3.1