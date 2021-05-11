// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kaniko

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/vol"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	"github.com/golang/glog"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logJob = log.Log.WithName("kaniko.job")

const (
	kanikoImgConst           = "kanikoImg"
	imagePushSecretNameConst = "imagePushSecretName"
	dockerRegCredVolumeName  = "reg-secret-volume"
)

// Job returns a kaniko job with mounted volumes
func Job(client *client.Client, api *wso2v1alpha1.API, controlConfigData map[string]string, kanikoArgs string, owner *[]metav1.OwnerReference, image registry.Image) (*batchv1.Job, error) {
	rootUserVal := int64(0)
	jobName := api.Name + "-kaniko"
	if api.Spec.UpdateTimeStamp != "" {
		jobName = jobName + "-" + api.Spec.UpdateTimeStamp
	}

	regConfig := registry.GetImageConfig(image)
	pushSecret := controlConfigData[imagePushSecretNameConst]
	if pushSecret != "" {
		regConfig.Volumes[0].VolumeSource.Secret.SecretName = pushSecret
	}
	AddVolumes(&regConfig.Volumes, &regConfig.VolumeMounts)

	kanikoImg := controlConfigData[kanikoImgConst]
	args := append([]string{
		"--dockerfile=/usr/wso2/dockerfile/Dockerfile",
		"--context=/usr/wso2/",
		"--destination=" + regConfig.ImagePath,
	}, regConfig.Args...)

	// if kaniko arguments are provided
	// read kaniko arguments and split them as they are read as a single string
	kanikoArguments := strings.Split(kanikoArgs, "\n")
	if kanikoArguments != nil {
		args = append(args, kanikoArguments...)
	}

	// Mount the user specified Config maps and secrets to mgw deploy volume
	userVol, userVolMount, envFromSources, err := vol.UserDeploymentVolume(client, api, vol.KanikoContext)
	if err != nil {
		return nil, err
	}
	*JobVolume = append(*JobVolume, userVol...)
	*JobVolumeMount = append(*JobVolumeMount, userVolMount...)

	var secretArray []corev1.LocalObjectReference
	secretArray = append(secretArray, regConfig.ImagePullSecrets...)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       api.Namespace,
			OwnerReferences: *owner,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      api.Name + "-job",
					Namespace: api.Namespace,
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         api.Name + "gen-container",
							Image:        kanikoImg,
							VolumeMounts: *JobVolumeMount,
							Args:         args,
							Env:          regConfig.Env,
							EnvFrom:      envFromSources,
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUserVal,
					},
					RestartPolicy:    "Never",
					Volumes:          *JobVolume,
					ImagePullSecrets: secretArray,
				},
			},
		},
	}, nil
}

// DeleteCompletedJob deletes completed kaniko jobs
func DeleteCompletedJob(namespace string) error {
	logJob.Info("Deleting completed kaniko job")
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Can't load in cluster config: %v", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Can't get client set: %v", err)
		return err
	}

	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &deletePolicy}
	//get list of exsisting jobs
	getListOfJobs, errGetJobs := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if len(getListOfJobs.Items) != 0 {
		for _, kanikoJob := range getListOfJobs.Items {
			if kanikoJob.Status.Succeeded > 0 {
				logJob.Info("Job "+kanikoJob.Name+" completed successfully", "Job.Namespace", kanikoJob.Namespace, "Job.Name", kanikoJob.Name)
				logJob.Info("Deleting job "+kanikoJob.Name, "Job.Namespace", kanikoJob.Namespace, "Job.Name", kanikoJob.Name)
				//deleting completed jobs
				errDelete := clientset.BatchV1().Jobs(kanikoJob.Namespace).Delete(context.TODO(), kanikoJob.Name, deleteOptions)
				if errDelete != nil {
					logJob.Error(errDelete, "error while deleting "+kanikoJob.Name+" job")
					return errDelete
				} else {
					logJob.Info("successfully deleted job "+kanikoJob.Name, "Job.Namespace", kanikoJob.Namespace, "Job.Name", kanikoJob.Name)
				}
			}
		}
	} else if errGetJobs != nil {
		logJob.Error(errGetJobs, "error retrieving jobs")
		return err
	}
	return nil
}
