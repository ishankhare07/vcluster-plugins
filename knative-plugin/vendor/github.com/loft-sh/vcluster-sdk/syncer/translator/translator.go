package translator

import (
	"github.com/loft-sh/vcluster-sdk/syncer/context"
	"github.com/loft-sh/vcluster-sdk/syncer/mapper"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// Translator is used to translate names as well as metadata between virtual and physical objects
type Translator interface {
	Resource() client.Object
	Name() string
	NameTranslator
	MetadataTranslator
}

// NameTranslator is used to convert virtual to physical names and vice versa
type NameTranslator interface {
	// IsManaged determines if a physical object is managed by the vcluster
	IsManaged(pObj client.Object) (bool, error)

	// VirtualToPhysical translates a virtual name to a physical name
	VirtualToPhysical(req types.NamespacedName, vObj client.Object) types.NamespacedName

	// PhysicalToVirtual translates a physical name to a virtual name
	PhysicalToVirtual(pObj client.Object) types.NamespacedName
}

// MetadataTranslator is used to convert metadata between virtual and physical objects and vice versa
type MetadataTranslator interface {
	// TranslateMetadata translates the object's metadata
	TranslateMetadata(vObj client.Object) client.Object

	// TranslateMetadataUpdate translates the object's metadata annotations and labels and determines
	// if they have changed between the physical and virtual object
	TranslateMetadataUpdate(vObj client.Object, pObj client.Object) (changed bool, annotations map[string]string, labels map[string]string)
}

// NamespacedTranslator provides some helper functions to ease sync down translation
type NamespacedTranslator interface {
	Translator
	mapper.Reverse

	// EventRecorder returns
	EventRecorder() record.EventRecorder

	// RegisterIndices registers the default indices for the syncer
	RegisterIndices(ctx *context.RegisterContext) error

	// SyncDownCreate creates the given pObj in the target namespace
	SyncDownCreate(ctx *context.SyncContext, vObj, pObj client.Object) (ctrl.Result, error)

	// SyncDownUpdate updates the given pObj (if not nil) in the target namespace
	SyncDownUpdate(ctx *context.SyncContext, vObj, pObj client.Object) (ctrl.Result, error)

	// AddReverseMapper adds additional indices and watchers for dependent objects
	AddReverseMapper(ctx *context.RegisterContext, obj client.Object, indexName string, indexer client.IndexerFunc, enqueuer handler.MapFunc)

	FindInNameCache(key types.NamespacedName) types.NamespacedName
	AddToNameCache(key, value types.NamespacedName)
}

// PhysicalNameTranslator transforms a virtual cluster name to a physical name
type PhysicalNameTranslator func(vName string, vObj client.Object) string
