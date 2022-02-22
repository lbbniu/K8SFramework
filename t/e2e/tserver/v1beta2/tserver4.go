package v1beta2

import (
	"context"
	"e2e/scaffold"
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	k8sCoreV1 "k8s.io/api/core/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	crdV1Beta2 "k8s.tars.io/api/crd/v1beta2"
	crdMeta "k8s.tars.io/api/meta"
	"strings"
)

var _ = ginkgo.Describe("try update tars server and check filed", func() {
	opts := &scaffold.Options{
		Name:      "default",
		K8SConfig: scaffold.GetK8SConfigFile(),
	}

	s := scaffold.NewScaffold(opts)

	var tsLayout *crdV1Beta2.TServer
	var Resource = "test-testserver"
	var App = "Test"
	var Server = "TestServer"
	var Template = "tt.cpp"
	var FirstObj = "FirstObj"
	var SecondObj = "SecondObj"

	ginkgo.BeforeEach(func() {
		ttLayout := &crdV1Beta2.TTemplate{
			ObjectMeta: k8sMetaV1.ObjectMeta{
				Name:      Template,
				Namespace: s.Namespace,
			},
			Spec: crdV1Beta2.TTemplateSpec{
				Content: "tt.cpp content",
				Parent:  Template,
			},
		}
		_, err := s.CRDClient.CrdV1beta2().TTemplates(s.Namespace).Create(context.TODO(), ttLayout, k8sMetaV1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err)

		tsLayout = &crdV1Beta2.TServer{
			ObjectMeta: k8sMetaV1.ObjectMeta{
				Name:      Resource,
				Namespace: s.Namespace,
			},
			Spec: crdV1Beta2.TServerSpec{
				App:       App,
				Server:    Server,
				SubType:   crdV1Beta2.TARS,
				Important: 5,
				Tars: &crdV1Beta2.TServerTars{
					Template:    Template,
					Profile:     "",
					AsyncThread: 3,
					Servants: []*crdV1Beta2.TServerServant{
						{
							Name:       FirstObj,
							Port:       10000,
							Thread:     3,
							Connection: 1000,
							Capacity:   1000,
							Timeout:    1000,
							IsTars:     true,
							IsTcp:      true,
						},
						{
							Name:       SecondObj,
							Port:       10001,
							Thread:     3,
							Connection: 1000,
							Capacity:   1000,
							Timeout:    1000,
							IsTars:     true,
							IsTcp:      true,
						},
					},
				},
				K8S: crdV1Beta2.TServerK8S{
					AbilityAffinity: crdV1Beta2.None,
					NodeSelector:    []k8sCoreV1.NodeSelectorRequirement{},
					ImagePullPolicy: k8sCoreV1.PullAlways,
					LauncherType:    crdV1Beta2.Background,
				},
			},
		}
		_, err = s.CRDClient.CrdV1beta2().TServers(s.Namespace).Create(context.TODO(), tsLayout, k8sMetaV1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err)
	})

	ginkgo.It("before update", func() {
		tserver, err := s.CRDClient.CrdV1beta2().TServers(s.Namespace).Get(context.TODO(), Resource, k8sMetaV1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.NotNil(ginkgo.GinkgoT(), tserver)

		expectedLabels := map[string]string{
			crdMeta.TServerAppLabel:  App,
			crdMeta.TServerNameLabel: Server,
			crdMeta.TemplateLabel:    Template,
			crdMeta.TSubTypeLabel:    string(crdV1Beta2.TARS),
		}
		assert.True(ginkgo.GinkgoT(), scaffold.CheckLeftInRight(expectedLabels, tserver.Labels))
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.Important)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.Tars)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.Tars.Servants)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.Tars.Template)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.Tars.AsyncThread)
		assert.Nil(ginkgo.GinkgoT(), tserver.Spec.Tars.Ports)
		assert.Nil(ginkgo.GinkgoT(), tserver.Spec.Normal)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.K8S)
		assert.False(ginkgo.GinkgoT(), tserver.Spec.K8S.HostIPC)
		assert.False(ginkgo.GinkgoT(), tserver.Spec.K8S.HostNetwork)
		assert.Nil(ginkgo.GinkgoT(), tserver.Spec.K8S.HostPorts)
		assert.False(ginkgo.GinkgoT(), tserver.Spec.K8S.DaemonSet)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.K8S.Replicas)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.K8S.NodeSelector)
		assert.NotNil(ginkgo.GinkgoT(), tserver.Spec.K8S.ImagePullPolicy)
	})

	ginkgo.Context("template", func() {
		var NewTemplate = "tt.xx"
		ginkgo.BeforeEach(func() {
			ttLayout := &crdV1Beta2.TTemplate{
				ObjectMeta: k8sMetaV1.ObjectMeta{
					Name:      NewTemplate,
					Namespace: s.Namespace,
				},
				Spec: crdV1Beta2.TTemplateSpec{
					Content: "tt.xx content",
					Parent:  NewTemplate,
				},
			}
			_, err := s.CRDClient.CrdV1beta2().TTemplates(s.Namespace).Create(context.TODO(), ttLayout, k8sMetaV1.CreateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)
		})

		ginkgo.It("not exist template", func() {
			jsonPatch := crdMeta.JsonPatch{
				{
					OP:    crdMeta.JsonPatchReplace,
					Path:  "/spec/tars/template",
					Value: "notexit",
				},
			}
			bs, _ := json.Marshal(jsonPatch)
			_, err := s.CRDClient.CrdV1beta2().TServers(s.Namespace).Patch(context.TODO(), Resource, patchTypes.JSONPatchType, bs, k8sMetaV1.PatchOptions{})
			assert.NotNil(ginkgo.GinkgoT(), err)
			assert.True(ginkgo.GinkgoT(), strings.Contains(err.Error(), "denied the request:"))
		})

		ginkgo.It("", func() {
			jsonPatch := crdMeta.JsonPatch{
				{
					OP:    crdMeta.JsonPatchReplace,
					Path:  "/spec/tars/template",
					Value: NewTemplate,
				},
			}
			bs, _ := json.Marshal(jsonPatch)
			tserver, err := s.CRDClient.CrdV1beta2().TServers(s.Namespace).Patch(context.TODO(), Resource, patchTypes.JSONPatchType, bs, k8sMetaV1.PatchOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.NotNil(ginkgo.GinkgoT(), tserver)

			expectedLabels := map[string]string{
				crdMeta.TServerAppLabel:  App,
				crdMeta.TServerNameLabel: Server,
				crdMeta.TemplateLabel:    NewTemplate,
				crdMeta.TSubTypeLabel:    string(crdV1Beta2.TARS),
			}
			assert.True(ginkgo.GinkgoT(), scaffold.CheckLeftInRight(expectedLabels, tserver.Labels))
		})
	})
})
