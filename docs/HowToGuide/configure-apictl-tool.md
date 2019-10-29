###Configure APICTL tool for API Operator

- Download API controller v3.0.0 for your operating system from - https://github.com/wso2/product-apim-tooling/releases/

- Extract the API controller distribution, and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

- You can find available operations using the below command.
    ```
    apictl --help
    ```
- This tool comes with the kubectl commands.
- However,by default API controller has disabled kubectl command. 
- When you are working with the API Operator, set the API Controller’s ***mode*** to Kubernetes as mentioned below.
    
    ```
    apictl set --mode k8s 
    ```
- If you wish to switch back to default mode, use the following command
    ```$xslt
    apictl set --mode default
    ```
<br />
