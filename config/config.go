package config

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Config config
type Config struct {
	SidecarImage string
	SidecarPort  int32
}

// Sidecar config
func (c *Config) Sidecar() corev1.Container {
	sidecar := corev1.Container{
		Name:  "kntool-sidecar",
		Image: c.SidecarImage,
		Ports: []corev1.ContainerPort{{
			ContainerPort: c.SidecarPort,
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
	return sidecar
}

var conf *Config

// Init init
func Init(c *Config) {
	conf = c
}

// GetConf get config
func GetConf() *Config {
	if conf == nil {
		panic("conf is nil")
	}
	return conf
}
