package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	cachev1alpha1 "github.com/Shehanir/k8s-securityOperator/security-operator/pkg/apis/cache/v1alpha1"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

)


type ValidationServerHandler struct {
}

func (vs *ValidationServerHandler) serve(w http.ResponseWriter, r *http.Request) {

	var body []byte
	var respmsg string
	var allow bool

	glog.Infof("Server running in webhook")

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	glog.Info("request received")

	if r.URL.Path != "/validate" {
		glog.Error("path validate not found")
		http.Error(w, "path validate not found", http.StatusBadRequest)
		return
	}

	//Unmarshal request body into a AdmissionReview struct
	arRequest := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &arRequest); err != nil {
		glog.Error("incorrect body")
		http.Error(w, "incorrect body", http.StatusBadRequest)
	}

	//Unmarshal byte data into a Security struct
	raw := arRequest.Request.Object.Raw
	security := cachev1alpha1.Security{}
	if err := json.Unmarshal(raw, &security); err != nil {
		glog.Error("error deserializing security object")
		return
	}

	fmt.Println(security.Name)

	secretgot := security.Spec.Credentials
	fmt.Println(secretgot)

	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Can't load in cluster config: %v", err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Can't get client set: %v", err)
		return
	}

	//fetch secret
	secret, err := clientset.CoreV1().Secrets("wso2-system").Get(secretgot,metav1.GetOptions{})
	if err != nil {

		glog.Errorf("secret: %v" + secretgot + "was not in the cluster", err)
		respmsg = "defined secret is not found"
		allow = false

		createArResponse(allow,respmsg,w)
		return
	}

	fmt.Println(secret.Name)
	respmsg = "successfully added security to the cluster"
	allow = true

	createArResponse(allow,respmsg,w)
	return

}

//create response
func createArResponse(allowed bool, message string , w http.ResponseWriter){

	arResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: allowed,
			Result: &metav1.Status{
				Message: message,
			},
		},
	}

	resp, err := json.Marshal(arResponse)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}

	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
