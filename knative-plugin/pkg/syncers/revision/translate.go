package revision

import (
	"github.com/loft-sh/vcluster-sdk/syncer/context"
	ksvcv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *revisionSyncer) ReverseTranslateMetadata(ctx *context.SyncContext, obj, parent client.Object) client.Object {
	rev := obj.(*ksvcv1.Revision)

	// reverse translate name and namespace
	namespacedName := r.PhysicalToVirtual(obj)

	rev.Name = namespacedName.Name
	rev.Namespace = namespacedName.Namespace

	// remove resourceVersion and uid
	rev.ObjectMeta.ResourceVersion = ""
	rev.ObjectMeta.UID = ""

	// reset owner references
	// TODO: find and set correct owner references
	// revert to config or route owner instead?
	var controller, bod *bool
	for _, owner := range rev.OwnerReferences {
		if owner.Kind == "Service" {
			controller = owner.Controller
			bod = owner.BlockOwnerDeletion
		}
	}

	parentKsvc := parent.(*ksvcv1.Service)

	rev.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         parentKsvc.APIVersion,
			Kind:               parentKsvc.Kind,
			Name:               parentKsvc.GetName(),
			UID:                parentKsvc.GetUID(),
			Controller:         controller,
			BlockOwnerDeletion: bod,
		},
	}

	return rev
}
