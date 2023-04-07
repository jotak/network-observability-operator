package watchers

import (
	"context"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const baseNamespace = "base-ns"
const otherNamespace = "other-ns"

var lokiCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "loki-ca",
		Namespace: baseNamespace,
	},
}
var lokiTLS = flowslatest.ClientTLS{
	Enable: true,
	CACert: flowslatest.CertificateReference{
		Type:     flowslatest.RefTypeConfigMap,
		Name:     lokiCA.Name,
		CertFile: "tls.crt",
	},
}
var otherLokiCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "other-loki-ca",
		Namespace: otherNamespace,
	},
}
var otherLokiTLS = flowslatest.ClientTLS{
	Enable: true,
	CACert: flowslatest.CertificateReference{
		Type:      flowslatest.RefTypeConfigMap,
		Name:      otherLokiCA.Name,
		Namespace: otherLokiCA.Namespace,
		CertFile:  "tls.crt",
	},
}
var kafkaCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kafka-ca",
		Namespace: baseNamespace,
	},
}
var kafkaUser = corev1.Secret{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kafka-user",
		Namespace: baseNamespace,
	},
}
var kafkaMTLS = flowslatest.ClientTLS{
	Enable: true,
	CACert: flowslatest.CertificateReference{
		Type:     flowslatest.RefTypeSecret,
		Name:     kafkaCA.Name,
		CertFile: "ca.crt",
	},
	UserCert: flowslatest.CertificateReference{
		Type:     flowslatest.RefTypeSecret,
		Name:     kafkaUser.Name,
		CertFile: "user.crt",
		CertKey:  "user.key",
	},
}
var kafkaSaslSecret = corev1.Secret{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kafka-sasl",
		Namespace: baseNamespace,
	},
	Data: map[string][]byte{
		"token": []byte("ssssaaaaassssslllll"),
	},
}
var kafkaSaslConfig = flowslatest.ConfigOrSecret{
	Type:     flowslatest.RefTypeSecret,
	Name:     kafkaSaslSecret.Name,
	Filename: "token",
}

func TestGenDigests(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockExisting(&lokiCA)
	clientMock.MockExisting(&kafkaCA)
	clientMock.MockExisting(&kafkaUser)
	clientMock.MockExisting(&kafkaSaslSecret)

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.ClientHelper{Client: &clientMock}

	dig1, dig2, err := watcher.ProcessMTLSCerts(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	assert.Equal("loki-ca", dig1)
	assert.Equal("", dig2)

	dig1, dig2, err = watcher.ProcessMTLSCerts(context.Background(), cl, &kafkaMTLS, baseNamespace)
	assert.NoError(err)
	assert.Equal("kafka-ca", dig1)
	assert.Equal("kafka-user", dig2)

	dig1, err = watcher.Process(context.Background(), cl, &kafkaSaslConfig, baseNamespace)
	assert.NoError(err)
	assert.Equal("kafka-sasl", dig1)
}

func TestNoCopy(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockExisting(&lokiCA)

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.ClientHelper{Client: &clientMock}

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertCalled(t, "Get", mock.Anything, types.NamespacedName{Name: lokiCA.Name, Namespace: lokiCA.Namespace}, mock.Anything)
	clientMock.AssertNotCalled(t, "Create")
	clientMock.AssertNotCalled(t, "Update")
}

func TestCopyCertificate(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockExisting(&otherLokiCA)
	clientMock.MockNonExisting(types.NamespacedName{Namespace: baseNamespace, Name: otherLokiCA.Name})
	clientMock.MockCreateUpdate()

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.ClientHelper{Client: &clientMock, SetControllerReference: func(o client.Object) error { return nil }}

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &otherLokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertCalled(t, "Get", mock.Anything, types.NamespacedName{Name: otherLokiCA.Name, Namespace: otherLokiCA.Namespace}, mock.Anything)
	clientMock.AssertCalled(t, "Get", mock.Anything, types.NamespacedName{Name: otherLokiCA.Name, Namespace: baseNamespace}, mock.Anything)
	clientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	clientMock.AssertNotCalled(t, "Update")
}

func TestUpdateCertificate(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockExisting(&otherLokiCA)
	copied := otherLokiCA
	copied.Namespace = baseNamespace
	clientMock.MockExisting(&copied)
	clientMock.MockCreateUpdate()

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.ClientHelper{Client: &clientMock, SetControllerReference: func(o client.Object) error { return nil }}

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &otherLokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertCalled(t, "Get", mock.Anything, types.NamespacedName{Name: otherLokiCA.Name, Namespace: otherLokiCA.Namespace}, mock.Anything)
	clientMock.AssertCalled(t, "Get", mock.Anything, types.NamespacedName{Name: otherLokiCA.Name, Namespace: baseNamespace}, mock.Anything)
	clientMock.AssertNotCalled(t, "Create")
	clientMock.AssertCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}
