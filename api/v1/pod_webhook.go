package v1

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodAnnotator struct {
	Client  client.Client
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,sideEffects=NoneOnDryRun,admissionReviewVersions=v1,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=xy.meteor.io

func (a *PodAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	//在 pod 中修改字段

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}
