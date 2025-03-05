package flp

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	informersName           = constants.FLPName + "-informers"
	informersConfigMap      = "flp-informers-config"
	informersTopic          = "informers"
	informersPromService    = informersName + "-prom"
	informersServiceMonitor = informersName + "-monitor"
)

type informersBuilder struct {
	info    *reconcilers.Instance
	desired *flowslatest.FlowCollectorSpec
	version string
	promTLS *flowslatest.CertificateReference
	volumes volumes.Builder
}

func newInformersBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec) (informersBuilder, error) {
	version := helper.ExtractVersion(info.Images[constants.ControllerBaseImageIndex])
	promTLS, err := getPromTLS(desired, informersPromService)
	if err != nil {
		return informersBuilder{}, err
	}
	return informersBuilder{
		info:    info,
		desired: desired,
		version: helper.MaxLabelLength(version),
		promTLS: promTLS,
	}, nil
}

func (b *informersBuilder) appLabel() map[string]string {
	return map[string]string{
		"app": informersName,
	}
}

func (b *informersBuilder) appVersionLabels() map[string]string {
	return map[string]string{
		"app":     informersName,
		"version": b.version,
	}
}

func (b *informersBuilder) deployment(annotations map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      informersName,
			Namespace: b.info.Namespace,
			Labels:    b.appVersionLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: b.appLabel(),
			},
			Template: b.podTemplate(annotations),
		},
	}
}

func (b *informersBuilder) podTemplate(annotations map[string]string) corev1.PodTemplateSpec {
	advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	var ports []corev1.ContainerPort

	if advancedConfig.ProfilePort != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: *advancedConfig.ProfilePort,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	volumeMounts := b.volumes.AppendMounts([]corev1.VolumeMount{{
		MountPath: configPath,
		Name:      configVolume,
	}})
	volumes := b.volumes.AppendVolumes([]corev1.Volume{{
		Name: configVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: informersConfigMap,
				},
			},
		},
	}})

	var envs []corev1.EnvVar
	envs = append(envs, constants.EnvNoHTTP2)

	container := corev1.Container{
		Name:            informersName,
		Image:           b.info.Images[constants.ControllerBaseImageIndex],
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Command:         []string{"/app/flp-informers"},
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
		SecurityContext: helper.ContainerDefaultSecurityContext(),
	}
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      b.appVersionLabels(),
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Volumes:            volumes,
			Containers:         []corev1.Container{container},
			ServiceAccountName: informersName,
		},
	}
}

func (b *informersBuilder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      informersName,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *informersBuilder) configMap() (*corev1.ConfigMap, string, error) {

	kafkaSpec := b.desired.Kafka
	metricsSettings := metricsSettings(b.desired, b.volumes, b.promTLS)
	cc := config.Informers{
		LogLevel: b.desired.Processor.LogLevel,
		KafkaConfig: api.EncodeKafka{
			Address: kafkaSpec.Address,
			Topic:   informersTopic,
			TLS:     getClientTLS(&b.desired.Kafka.TLS, "kafka-cert", &b.volumes),
			SASL:    getSASL(&b.desired.Kafka.SASL, "kafka-ingest", &b.volumes),
		},
		MetricsSettings: metricsSettings,
	}
	advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	if advancedConfig.ProfilePort != nil {
		cc.PProfPort = *advancedConfig.ProfilePort
	}

	bs, err := yaml.Marshal(cc)
	if err != nil {
		return nil, "", err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      informersConfigMap,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
		Data: map[string]string{
			configFile: string(bs),
		},
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write(bs)
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest, nil
}

func (b *informersBuilder) promService() *corev1.Service {
	port := helper.GetFlowCollectorMetricsPort(b.desired)
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      informersPromService,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
		Spec: corev1.ServiceSpec{
			Selector: b.appLabel(),
			Ports: []corev1.ServicePort{{
				Name:     prometheusServiceName,
				Port:     port,
				Protocol: corev1.ProtocolTCP,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt32(port),
			}},
		},
	}
	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSAuto {
		svc.ObjectMeta.Annotations = map[string]string{
			constants.OpenShiftCertificateAnnotation: informersPromService,
		}
	}
	return &svc
}

func (b *informersBuilder) serviceMonitor() *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", informersPromService, b.info.Namespace)
	scheme, smTLS := helper.GetServiceMonitorTLSConfig(&b.desired.Processor.Metrics.Server.TLS, serverName, b.info.IsDownstream)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      informersServiceMonitor,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:      prometheusServiceName,
					Interval:  "30s",
					Scheme:    scheme,
					TLSConfig: smTLS,
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{b.info.Namespace},
			},
			Selector: metav1.LabelSelector{MatchLabels: b.appLabel()},
		},
	}
}
