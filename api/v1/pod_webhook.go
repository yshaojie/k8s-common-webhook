package v1

import (
	"context"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodAnnotator struct {
	Client  client.Client
	decoder *admission.Decoder
}

var (
	log           = ctrl.Log.WithName("webhook")
	commonEnvVars = []corev1.EnvVar{{
		Name: "K8S_WORKER_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "spec.nodeName",
			},
		},
	}, {
		Name: "K8S_POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}, {
		Name: "K8S_POD_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	}, {
		Name: "K8S_POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	}, {
		Name: "K8S_WORKER_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	}}
)

//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,sideEffects=NoneOnDryRun,admissionReviewVersions=v1,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=xy.meteor.io

func (a *PodAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Kind.Kind != "Pod" {
		return admission.Allowed("not a pod,skip")
	}

	switch req.Operation {
	case admissionv1.Create:
	case admissionv1.Update:
		return admission.Allowed("skip")
	default:
		return admission.Allowed("skip")
	}
	//在 pod 中修改字段
	mutatePod(pod)
	marshaledPod, err := json.Marshal(pod)
	log.Info(string(marshaledPod))
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func mutatePod(pod *corev1.Pod) error {
	containers := pod.Spec.Containers
	fillCommonEnvVars(containers)
	initContainers := pod.Spec.InitContainers
	fillCommonEnvVars(initContainers)
	config := pod.Spec.DNSConfig
	if config == nil {
		pod.Spec.DNSConfig = &corev1.PodDNSConfig{}
		config = pod.Spec.DNSConfig
	}
	dnsConfigOptions := config.Options
	if dnsConfigOptions == nil {
		config.Options = []corev1.PodDNSConfigOption{}
		dnsConfigOptions = config.Options
	}

	dnsConfigOption := corev1.PodDNSConfigOption{Name: "single-request-reopen"}
	if !hasDnsConfigOptions(dnsConfigOptions, dnsConfigOption) {
		dnsConfigOptions = append(dnsConfigOptions, dnsConfigOption)
	}
	config.Options = dnsConfigOptions
	return nil
}

func hasDnsConfigOptions(options []corev1.PodDNSConfigOption, targetOption corev1.PodDNSConfigOption) bool {
	for _, podDNSConfigOption := range options {
		if podDNSConfigOption.Name == targetOption.Name && podDNSConfigOption.Value == targetOption.Value {
			return true
		}
	}

	return false
}

func fillCommonEnvVars(containers []corev1.Container) error {
	for i, container := range containers {
		envVars := container.Env
		for _, commonEnvVar := range commonEnvVars {
			if hasEnvVar(container, commonEnvVar) {
				continue
			}
			envVars = append(envVars, commonEnvVar)
		}
		containers[i].Env = envVars
	}

	return nil
}

func hasEnvVar(container corev1.Container, checkEnvVar corev1.EnvVar) bool {
	for _, envVar := range container.Env {
		if envVar.Name == checkEnvVar.Name {
			return true
		}
	}
	return false
}

func (a *PodAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
