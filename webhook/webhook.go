package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	admissionV1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	// 定义基础工具
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	_ = runtime.ObjectDefaulter(runtimeScheme)
)

var (
	// TODO ignored namespaces
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

// HandlerMutate handler mutate
func HandlerMutate(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if contentType != "application/json" {
		logrus.Errorf("Content-type=%s, expect application/json", contentType)
		http.Error(c.Writer, "invalid Content-Type, expect `application/json`", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var admissionResponse *admissionV1.AdmissionResponse
	admissionReview := new(admissionV1.AdmissionReview)

	if _, _, err := deserializer.Decode(body, nil, admissionReview); err != nil {
		logrus.Errorf("Can't decode body: %v", err)

		admissionResponse = &admissionV1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = mutate(admissionReview)
	}

	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if admissionReview.Request != nil {
			admissionReview.Response.UID = admissionReview.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		logrus.Errorf("Can't encode response: %v", err)
		http.Error(c.Writer, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}

	if _, err := c.Writer.Write(resp); err != nil {
		http.Error(c.Writer, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func mutate(admissionReview *admissionV1.AdmissionReview) *admissionV1.AdmissionResponse {
	req := admissionReview.Request

	logrus.WithField("kind", req.Kind).
		WithField("namespace", req.Namespace).
		WithField("uid", req.UID).
		WithField("operation", req.Operation).
		WithField("userInfo", req.UserInfo)

	switch req.Kind.Kind {
	case "Pod":
		pod := new(corev1.Pod)
		if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
			logrus.Errorf("could not unmarshal raw object: %v", err)
			return &admissionV1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}

		patch, err := createPatch(pod)
		if err != nil {
			logrus.Error(err)
			return &admissionV1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}

		return &admissionV1.AdmissionResponse{
			Allowed: true,
			Patch:   patch,
			PatchType: func() *admissionV1.PatchType {
				pt := admissionV1.PatchTypeJSONPatch
				return &pt
			}(),
		}
	default:
		logrus.Errorf("unsupported resource type")
		return &admissionV1.AdmissionResponse{
			Result: &metav1.Status{
				Message: "Unsupported resource type",
			},
		}
	}
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// create mutation patch for resoures
func createPatch(pod *corev1.Pod) ([]byte, error) {
	var patch []patchOperation

	patch = append(patch, addSidecar(pod.Spec.Containers)...)

	return json.Marshal(patch)
}

func addSidecar(target []corev1.Container) (patch []patchOperation) {
	sidecar := corev1.Container{
		Name:  "kntool-sidecar",
		Image: "zhaihuailou/kntool-sidecar",
		Ports: []corev1.ContainerPort{{
			ContainerPort: 2332,
		}},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"NET_ADMIN",
				},
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.2"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.2"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
		},
		ImagePullPolicy: "IfNotPresent",
	}

	var find bool

	for i, container := range target {
		if container.Name == "kntool-sidercar" {
			target[i] = sidecar
			find = true
		}
	}

	if !find {
		target = append(target, sidecar)
	}

	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/spec/containers",
		Value: target,
	})

	return patch
}
