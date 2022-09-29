package tracing

import (
	"context"
	"fmt"

	addonclient "open-cluster-management.io/api/client/addon/clientset/versioned"
	addoninformerv1alpha1 "open-cluster-management.io/api/client/addon/informers/externalversions/addon/v1alpha1"
	addonlisterv1alpha1 "open-cluster-management.io/api/client/addon/listers/addon/v1alpha1"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/apimachinery/pkg/labels"

	"go.opentelemetry.io/otel"
)

var enabledHub bool = false
var tpHub func()

type otelCollectorAddonControllerHub struct {
	servicename string
	addOnClient addonclient.Interface
	addOnLister addonlisterv1alpha1.ClusterManagementAddOnLister
}

func NewOtelCollectorAddonControllerHub(servicename string,
	addOnClient addonclient.Interface,
	addOnInformer addoninformerv1alpha1.ClusterManagementAddOnInformer,
	recorder events.Recorder) factory.Controller {
	c := &otelCollectorAddonControllerHub{
		servicename: servicename,
		addOnClient: addOnClient,
		addOnLister: addOnInformer.Lister(),
	}

	return factory.New().
		WithInformers(addOnInformer.Informer()).
		WithSync(c.sync).
		ToController("OtelCollectorAddonControllerHub", recorder)
}

func (c *otelCollectorAddonControllerHub) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	_, span := otel.Tracer("otelCollectorAddonController").Start(ctx, "Addon - otelCollectorAddonController")
	defer span.End()

	queueKey := syncCtx.QueueKey()
	if queueKey == factory.DefaultQueueKey {
		addOns,err := c.addOnLister.List(labels.Everything())
		if err != nil {
			return err
		}
		for _, addOn := range addOns {
			if addOn.Name == "otel-collector" && !enabledHub {
				tpHub = Initialize(c.servicename,"collector-service.open-cluster-management-addon.svc.cluster.local:4317")
				fmt.Println("---------Tracing Enabled --------------")
				enabledHub = true
				return nil
			}
			if addOn.Name == "otel-collector" && enabledHub {
				fmt.Println("---------Tracing already enabled --------------")
				//Return no need to do anything
				return nil
			}
		}
		if enabledHub {
			//closing the tracer
			fmt.Println("---------Disabling Tracing--------------")
			tpHub()
			enabledHub = false
			return nil
		}
	}
	return nil
}
