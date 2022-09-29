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

var enabled bool = false
var tp func()

type otelCollectorAddonController struct {
	servicename string
	clusterName string
	addOnClient addonclient.Interface
	addOnLister addonlisterv1alpha1.ManagedClusterAddOnLister
}

func NewOtelCollectorAddonController(servicename string,
	clusterName string,
	addOnClient addonclient.Interface,
	addOnInformer addoninformerv1alpha1.ManagedClusterAddOnInformer,
	recorder events.Recorder) factory.Controller {
	c := &otelCollectorAddonController{
		servicename: servicename,
		clusterName: clusterName,
		addOnClient: addOnClient,
		addOnLister: addOnInformer.Lister(),
	}

	return factory.New().
		WithInformers(addOnInformer.Informer()).
		WithSync(c.sync).
		ToController("OtelCollectorAddonController", recorder)
}

func (c *otelCollectorAddonController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	_, span := otel.Tracer("otelCollectorAddonController").Start(ctx, "Addon - otelCollectorAddonController")
	defer span.End()

	queueKey := syncCtx.QueueKey()
	if queueKey == factory.DefaultQueueKey {
		addOns, err := c.addOnLister.ManagedClusterAddOns(c.clusterName).List(labels.Everything())
		if err != nil {
			return err
		}
		for _, addOn := range addOns {
			if addOn.Name == "otel-collector" && !enabled {
				tp=Initialize(c.servicename, "collector-service.open-cluster-management-otel-collector.svc.cluster.local:4317")
				fmt.Println("---------Tracing Enabled--------------")
				enabled = true
				return nil
			}
			if addOn.Name == "otel-collector" && enabled {
				//Return no need to do anything
				return nil
			}
		}
		if enabled {
			//closing the tracer
			fmt.Println("---------Disabling Tracing--------------")
			tp()
			enabled=false
			return nil
		}
	}
	return nil
}
