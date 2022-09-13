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

	// check spec
	if !equality.Semantic.DeepEqual(pObj.Spec.Traffic, vObj.Spec.Traffic) {
		klog.Infof("spec.traffic for vKsvc %s:%s, is out of sync", vObj.Namespace, vObj.Name)
		updated = newIfNil(updated, vObj)

		// TODO: have more fine grained backwards sync
		// we should not allow physical object traffic to modify
		// revision names or traffic percent in virtual object
		updated.Spec.Traffic = pObj.Spec.Traffic
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
