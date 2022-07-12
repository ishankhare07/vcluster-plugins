package configurations

import (
	"github.com/loft-sh/vcluster-sdk/syncer"
	"github.com/loft-sh/vcluster-sdk/syncer/context"
	"github.com/loft-sh/vcluster-sdk/syncer/translator"
	"github.com/loft-sh/vcluster-sdk/translate"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	ksvcv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func New(ctx *context.RegisterContext) syncer.Syncer {
	return &kconfigSyncer{
		NamespacedTranslator: translator.NewNamespacedTranslator(ctx, "configuration", &ksvcv1.Configuration{}),
	}
}

type kconfigSyncer struct {
	translator.NamespacedTranslator
}

func (k *kconfigSyncer) Init(ctx *context.RegisterContext) error {
	return translate.EnsureCRDFromPhysicalCluster(ctx.Context,
		ctx.PhysicalManager.GetConfig(),
		ctx.VirtualManager.GetConfig(),
		ksvcv1.SchemeGroupVersion.WithKind("Configuration"))
}

// SyncDown defines the action that should be taken by the syncer if a virtual cluster object
// exists, but has no corresponding physical cluster object yet. Typically, the physical cluster
// object would get synced down from the virtual cluster to the host cluster in this scenario.
func (k *kconfigSyncer) SyncDown(ctx *context.SyncContext, vObj client.Object) (ctrl.Result, error) {
	klog.Info("SyncDown called for ", vObj.GetName())

	klog.Infof("Deleting virtual Config Object %s because physical no longer exists", vObj.GetName())
	err := ctx.VirtualClient.Delete(ctx.Context, vObj)
	if err != nil {
		klog.Infof("Error deleting virtual Config object: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (k *kconfigSyncer) Sync(ctx *context.SyncContext, pObj, vObj client.Object) (ctrl.Result, error) {
	klog.Infof("Sync called for Configuration %s : %s", pObj.GetName(), vObj.GetName())

	pConfig := pObj.(*ksvcv1.Configuration)
	vConfig := vObj.(*ksvcv1.Configuration)

	if !equality.Semantic.DeepEqual(vConfig.Status, pConfig.Status) {
		newConfig := vConfig.DeepCopy()
		newConfig.Status = pConfig.Status
		klog.Infof("Update virtual kconfig %s:%s, because status is out of sync", vConfig.Namespace, vConfig.Name)
		err := ctx.VirtualClient.Status().Update(ctx.Context, newConfig)
		if err != nil {
			klog.Errorf("Error updating virtual kconfig status for %s:%s, %v", vConfig.Namespace, vConfig.Name, err)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (k *kconfigSyncer) SyncUp(ctx *context.SyncContext, pObj client.Object) (ctrl.Result, error) {
	klog.Info("SyncUp called for ", pObj.GetName())

	return k.SyncUpCreate(ctx, pObj)
}

func (k *kconfigSyncer) SyncUpCreate(ctx *context.SyncContext, pObj client.Object) (ctrl.Result, error) {
	klog.Infof("SyncUpCreate called for %s:%s", pObj.GetName(), pObj.GetNamespace())
	klog.Info("reverse name should be ", k.PhysicalToVirtual(pObj))

	// k.TranslateMetadata()
	pName := k.PhysicalToVirtual(pObj)
	pConfig := pObj.(*ksvcv1.Configuration)
	pConfig.ObjectMeta.Name = pName.Name
	pConfig.ObjectMeta.Namespace = pName.Namespace

	// remove resourceVersion and uid
	pConfig.ObjectMeta.ResourceVersion = ""
	pConfig.ObjectMeta.UID = ""

	// for i, ownerRef := range pConfig.OwnerReferences {
	// 	pConfig.OwnerReferences[i].Name =
	// }

	pConfig.OwnerReferences = []metav1.OwnerReference{}

	err := ctx.VirtualClient.Create(ctx.Context, pObj)
	if err != nil {
		k.NamespacedTranslator.EventRecorder().Eventf(pObj, "Warning", "SyncError", "Error syncing to virtual cluster: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
