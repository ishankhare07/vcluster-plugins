package ksvc

import (
	"github.com/loft-sh/vcluster-sdk/translate"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog"
	ksvcv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *ksvcSyncer) translate(vObj client.Object) *ksvcv1.Service {
	pObj := k.TranslateMetadata(vObj).(*ksvcv1.Service)
	vKsvc := vObj.(*ksvcv1.Service)

	pObj.Spec = *rewriteSpec(&vKsvc.Spec, vKsvc.Namespace)

	return pObj
}

func (k *ksvcSyncer) translateUpdate(pObj, vObj *ksvcv1.Service) *ksvcv1.Service {
	if equality.Semantic.DeepEqual(pObj, vObj) {
		klog.Infof("no diff found, returning")
		return nil
	}

	var newPKsvc *ksvcv1.Service

	// check for configuration fields
	newPKsvc = updateConfigurationSpec(newPKsvc, pObj, vObj)
	newPKsvc = updateTrafficSpec(newPKsvc, pObj, vObj)

	return newPKsvc
}

func (k *ksvcSyncer) translateUpdateBackwards(pObj, vObj *ksvcv1.Service) *ksvcv1.Service {
	var updated *ksvcv1.Service

	// check annotations
	if !equality.Semantic.DeepEqual(pObj.ObjectMeta.Annotations, vObj.ObjectMeta.Annotations) {
		klog.Infof("annotations for vKsvc %s:%s, is out of sync", vObj.Namespace, vObj.Name)
		updated = newIfNil(updated, vObj)

		updated.ObjectMeta.Annotations = pObj.ObjectMeta.Annotations
	}

	// check spec
	if !equality.Semantic.DeepEqual(pObj.Spec.Traffic, vObj.Spec.Traffic) {
		klog.Infof("spec.traffic for vKsvc %s:%s, is out of sync", vObj.Namespace, vObj.Name)
		updated = newIfNil(updated, vObj)

		// TODO: have more fine grained backwards sync
		// we should not allow physical object traffic to modify
		// revision names or traffic percent in virtual object
		updated.Spec.Traffic = pObj.Spec.Traffic
	}

	// check RevisionSpec
	if !equality.Semantic.DeepEqual(pObj.Spec.Template.Spec, vObj.Spec.Template.Spec) {
		if vObj.Spec.Template.Spec.PodSpec.Containers[0].Name == "" {
			updated = newIfNil(updated, vObj)
			updated.Spec.Template.Spec.PodSpec.Containers[0].Name = pObj.Spec.Template.Spec.PodSpec.Containers[0].Name
		}

		if vObj.Spec.Template.Spec.PodSpec.EnableServiceLinks == nil {
			updated = newIfNil(updated, vObj)
			updated.Spec.Template.Spec.PodSpec.EnableServiceLinks = pObj.Spec.Template.Spec.PodSpec.EnableServiceLinks
		}

		if vObj.Spec.Template.Spec.ContainerConcurrency == nil {
			updated = newIfNil(updated, vObj)
			updated.Spec.Template.Spec.ContainerConcurrency = pObj.Spec.Template.Spec.ContainerConcurrency
		}

		if vObj.Spec.Template.Spec.TimeoutSeconds == nil {
			updated = newIfNil(updated, vObj)
			updated.Spec.Template.Spec.TimeoutSeconds = pObj.Spec.Template.Spec.TimeoutSeconds
		}

		if vObj.Spec.Template.Spec.ResponseStartTimeoutSeconds == nil {
			updated = newIfNil(updated, vObj)
			updated.Spec.Template.Spec.ResponseStartTimeoutSeconds = pObj.Spec.Template.Spec.ResponseStartTimeoutSeconds
		}

		if vObj.Spec.Template.Spec.IdleTimeoutSeconds == nil {
			updated = newIfNil(updated, vObj)
			updated.Spec.Template.Spec.IdleTimeoutSeconds = pObj.Spec.Template.Spec.IdleTimeoutSeconds
		}
	}

	return updated

}

func rewriteSpec(vObjSpec *ksvcv1.ServiceSpec, namespace string) *ksvcv1.ServiceSpec {
	vObjSpec = vObjSpec.DeepCopy()

	klog.Info("template name: ", vObjSpec.ConfigurationSpec.Template.Name)
	if vObjSpec.ConfigurationSpec.Template.Name != "" {
		vObjSpec.ConfigurationSpec.Template.Name = translate.PhysicalName(vObjSpec.ConfigurationSpec.Template.Name, namespace)
	}

	return vObjSpec
}

func newIfNil(updated, obj *ksvcv1.Service) *ksvcv1.Service {
	if updated == nil {
		return obj.DeepCopy()
	}

	return updated
}

func updateConfigurationSpec(newPKsvc, pObj, vObj *ksvcv1.Service) *ksvcv1.Service {
	klog.Info("checking diff in configuration spec")
	klog.Infof("vImage: ", vObj.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image)
	klog.Infof("pImage: ", pObj.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image)
	if !equality.Semantic.DeepEqual(
		vObj.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image,
		pObj.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image) {

		newPKsvc = newIfNil(newPKsvc, pObj)

		klog.Infof("image different for vKsvc %s:%s, syncing down", vObj.Namespace, vObj.Name)
		newPKsvc.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image = vObj.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image
	}

	// check diff in containerConcurrency
	if vObj.Spec.ConfigurationSpec.Template.Spec.ContainerConcurrency != nil {
		if !equality.Semantic.DeepEqual(
			vObj.Spec.ConfigurationSpec.Template.Spec.ContainerConcurrency,
			pObj.Spec.ConfigurationSpec.Template.Spec.ContainerConcurrency) {

			newPKsvc = newIfNil(newPKsvc, pObj)

			klog.Infof("containerConcurrency different for vKsvc %s:%s, syncing down", vObj.Namespace, vObj.Name)
			newPKsvc.Spec.ConfigurationSpec.Template.Spec.ContainerConcurrency = vObj.Spec.ConfigurationSpec.Template.Spec.ContainerConcurrency
		}
	}

	// check diff in timeoutSeconds
	if vObj.Spec.ConfigurationSpec.Template.Spec.TimeoutSeconds != nil {
		if !equality.Semantic.DeepEqual(
			vObj.Spec.ConfigurationSpec.Template.Spec.TimeoutSeconds,
			pObj.Spec.ConfigurationSpec.Template.Spec.TimeoutSeconds) {

			newPKsvc = newIfNil(newPKsvc, pObj)

			klog.Infof("timeoutSeconds different for vKsvc %s:%s, syncing down", vObj.Namespace, vObj.Name)
			newPKsvc.Spec.ConfigurationSpec.Template.Spec.TimeoutSeconds = vObj.Spec.ConfigurationSpec.Template.Spec.TimeoutSeconds
		}
	}

	return newPKsvc
}

func updateTrafficSpec(newPKsvc, pObj, vObj *ksvcv1.Service) *ksvcv1.Service {
	if vObj.Spec.RouteSpec.Traffic == nil {
		// vksvc not set by user, do not sync down
		// and let controller set defaults to the traffic field
		return newPKsvc
	}

	if !equality.Semantic.DeepEqual(
		vObj.Spec.RouteSpec.Traffic,
		pObj.Spec.RouteSpec.Traffic) {

		newPKsvc = newIfNil(newPKsvc, pObj)
		klog.Infof("traffic spec different for vKsvc %s:%s, syncing down", vObj.Namespace, vObj.Name)
		newPKsvc.Spec.RouteSpec.Traffic = vObj.Spec.RouteSpec.Traffic
	}

	return newPKsvc
}
