package flp

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/resources"
)

type informersReconciler struct {
	*reconcilers.Instance
	deployment     *appsv1.Deployment
	promService    *corev1.Service
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBinding    *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
}

func newInformersReconciler(cmn *reconcilers.Instance) *informersReconciler {
	rec := informersReconciler{
		Instance:       cmn,
		deployment:     cmn.Managed.NewDeployment(informersName),
		promService:    cmn.Managed.NewService(informersPromService),
		serviceAccount: cmn.Managed.NewServiceAccount(informersName),
		configMap:      cmn.Managed.NewConfigMap(informersConfigMap),
		roleBinding:    cmn.Managed.NewCRB(resources.GetClusterRoleBindingName(informersName, constants.FLPInformersRole)),
	}
	if cmn.ClusterInfo.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(serviceMonitorName(ConfKafkaTransformer))
	}
	return &rec
}

func (r *informersReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithName("informers")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *informersReconciler) cleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

func (r *informersReconciler) getStatus() *status.Instance {
	return &r.Status
}

func (r *informersReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector, _ *metricslatest.FlowMetricList, _ []flowslatest.SubnetLabel) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if !helper.UseSharedInformers(&desired.Spec) {
		r.Status.SetUnused("Shared informers is not enabled")
		r.Managed.TryDeleteAll(ctx)
		return nil
	}

	r.Status.SetReady() // will be overidden if necessary, as error or pending

	builder, err := newInformersBuilder(r.Instance, &desired.Spec)
	if err != nil {
		return err
	}

	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return err
	}
	annotations := map[string]string{constants.PodConfigurationDigest: configDigest}
	if err := reconcilers.ReconcileConfigMapManaged(ctx, r.Instance, r.configMap, newCM); err != nil {
		return err
	}

	if err := r.reconcilePermissions(ctx, &builder); err != nil {
		return err
	}

	err = r.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return err
	}

	// Watch for Kafka certificate if necessary; need to restart pods in case of cert rotation
	if err = annotateKafkaCerts(ctx, r.Common, &desired.Spec.Kafka, "kafka", annotations); err != nil {
		return err
	}

	// Watch for monitoring caCert
	if err = reconcileMonitoringCerts(ctx, r.Common, &desired.Spec.Processor.Metrics.Server.TLS, r.Namespace); err != nil {
		return err
	}

	return r.reconcileDeployment(ctx, &builder, annotations)
}

func (r *informersReconciler) reconcileDeployment(ctx context.Context, builder *informersBuilder, annotations map[string]string) error {
	report := helper.NewChangeReport("FLP Informers deployment")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDeployment(
		ctx,
		r.Instance,
		r.deployment,
		builder.deployment(annotations),
		constants.FLPName,
		1,
		nil,
		&report,
	)
}

func (r *informersReconciler) reconcilePrometheusService(ctx context.Context, builder *informersBuilder) error {
	report := helper.NewChangeReport("FLP Informers prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.ClusterInfo.HasSvcMonitor() {
		serviceMonitor := builder.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *informersReconciler) reconcilePermissions(ctx context.Context, builder *informersBuilder) error {
	if !r.Managed.Exists(r.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	binding := resources.GetClusterRoleBinding(r.Namespace, informersName, informersName, constants.FLPInformersRole)
	return r.ReconcileClusterRoleBinding(ctx, binding)
}
