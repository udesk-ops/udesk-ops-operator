package handlers_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/constants"
	"udesk.cn/ops/internal/server/handlers"
	"udesk.cn/ops/internal/types"
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func TestApprovalHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Approval Handlers Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "bin", "k8s", "1.33.0-linux-amd64"),
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = opsv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("CommonApprovalProcessor", func() {
	var (
		processor *handlers.CommonApprovalProcessor
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		processor = handlers.NewCommonApprovalProcessor(k8sClient)
	})

	Describe("ProcessApprovalRequest", func() {
		Context("when processing AlertScale approval", func() {
			var alertScale *opsv1beta1.AlertScale

			BeforeEach(func() {
				alertScale = &opsv1beta1.AlertScale{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alertscale",
						Namespace: "default",
					},
					Spec: opsv1beta1.AlertScaleSpec{
						ScaleReason: "High CPU usage",
						ScaleTarget: opsv1beta1.ScaleTarget{
							Kind: "Deployment",
							Name: "test-deployment",
						},
						ScaleDuration: "10m",
					},
					Status: opsv1beta1.AlertScaleStatus{
						ScaleStatus: opsv1beta1.ScaleStatus{
							Status: types.ScaleStatusApprovaling,
						},
					},
				}
				Expect(k8sClient.Create(ctx, alertScale)).To(Succeed())

				// Update status separately - required in test environment
				alertScale.Status.ScaleStatus.Status = types.ScaleStatusApprovaling
				Expect(k8sClient.Status().Update(ctx, alertScale)).To(Succeed())
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, alertScale)).To(Succeed())
			})

			It("should approve AlertScale successfully", func() {
				// Create approval request
				req := handlers.CommonApprovalRequest{
					Approver: "test-approver",
					Reason:   "Looks good",
					Comment:  "Test comment",
				}

				// Process approval using the specialized method
				resourceKey := client.ObjectKey{Namespace: alertScale.Namespace, Name: alertScale.Name}
				err := processor.ProcessAlertScaleApproval(ctx, resourceKey, "approve", req)
				Expect(err).ToNot(HaveOccurred())

				// Verify annotations were set
				var updated opsv1beta1.AlertScale
				Expect(k8sClient.Get(ctx, resourceKey, &updated)).To(Succeed())

				Expect(updated.Annotations).To(HaveKey(constants.ApprovalDecisionAnnotation))
				Expect(updated.Annotations[constants.ApprovalDecisionAnnotation]).To(Equal("approve"))
				Expect(updated.Annotations[constants.ApprovalOperatorAnnotation]).To(Equal("test-approver"))
				Expect(updated.Annotations[constants.ApprovalReasonAnnotation]).To(Equal("Looks good"))
				Expect(updated.Annotations[constants.ApprovalCommentAnnotation]).To(Equal("Test comment"))
				Expect(updated.Annotations[constants.ApprovalProcessingAnnotation]).To(Equal("pending"))
			})

			It("should reject AlertScale successfully", func() {
				// Create rejection request
				req := handlers.CommonApprovalRequest{
					Approver: "test-rejector",
					Reason:   "Not safe",
				}

				// Process rejection using the specialized method
				resourceKey := client.ObjectKey{Namespace: alertScale.Namespace, Name: alertScale.Name}
				err := processor.ProcessAlertScaleApproval(ctx, resourceKey, "reject", req)
				Expect(err).ToNot(HaveOccurred())

				// Verify annotations were set
				var updated opsv1beta1.AlertScale
				Expect(k8sClient.Get(ctx, resourceKey, &updated)).To(Succeed())

				Expect(updated.Annotations).To(HaveKey(constants.ApprovalDecisionAnnotation))
				Expect(updated.Annotations[constants.ApprovalDecisionAnnotation]).To(Equal("reject"))
				Expect(updated.Annotations[constants.ApprovalOperatorAnnotation]).To(Equal("test-rejector"))
				Expect(updated.Annotations[constants.ApprovalReasonAnnotation]).To(Equal("Not safe"))
			})
		})

		Context("when processing PodRebalance approval", func() {
			var podRebalance *opsv1beta1.PodRebalance

			BeforeEach(func() {
				podRebalance = &opsv1beta1.PodRebalance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-podrebalance",
						Namespace: "default",
					},
					Spec: opsv1beta1.PodRebalanceSpec{
						Namespace: "default",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
						Strategy: opsv1beta1.PodRebalanceStrategy{
							Type: "NodeBalance",
						},
					},
					Status: opsv1beta1.PodRebalanceStatus{
						Status: types.RebalanceStatusApprovaling,
					},
				}
				Expect(k8sClient.Create(ctx, podRebalance)).To(Succeed())

				// Update status separately - required in test environment
				podRebalance.Status.Status = types.RebalanceStatusApprovaling
				Expect(k8sClient.Status().Update(ctx, podRebalance)).To(Succeed())
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, podRebalance)).To(Succeed())
			})

			It("should approve PodRebalance successfully", func() {
				// Create approval request
				req := handlers.CommonApprovalRequest{
					Approver: "test-approver",
					Reason:   "Balancing needed",
				}

				// Process approval using the specialized method
				resourceKey := client.ObjectKey{Namespace: podRebalance.Namespace, Name: podRebalance.Name}
				err := processor.ProcessPodRebalanceApproval(ctx, resourceKey, "approve", req)
				Expect(err).ToNot(HaveOccurred())

				// Verify annotations were set
				var updated opsv1beta1.PodRebalance
				Expect(k8sClient.Get(ctx, resourceKey, &updated)).To(Succeed())

				Expect(updated.Annotations).To(HaveKey(constants.ApprovalDecisionAnnotation))
				Expect(updated.Annotations[constants.ApprovalDecisionAnnotation]).To(Equal("approve"))
				Expect(updated.Annotations[constants.ApprovalOperatorAnnotation]).To(Equal("test-approver"))
				Expect(updated.Annotations[constants.ApprovalReasonAnnotation]).To(Equal("Balancing needed"))
			})
		})

		Context("when validation fails", func() {
			It("should return error for invalid action", func() {
				alertScale := &opsv1beta1.AlertScale{}
				adapter := handlers.NewAlertScaleApprovalAdapter(alertScale)
				req := handlers.CommonApprovalRequest{
					Approver: "test-approver",
					Reason:   "test",
				}

				resourceKey := client.ObjectKey{Namespace: "default", Name: "test"}
				err := processor.ProcessApprovalRequest(ctx, resourceKey, adapter, "invalid", req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid action"))
			})

			It("should return error for missing approver", func() {
				alertScale := &opsv1beta1.AlertScale{}
				adapter := handlers.NewAlertScaleApprovalAdapter(alertScale)
				req := handlers.CommonApprovalRequest{
					Reason: "test",
				}

				resourceKey := client.ObjectKey{Namespace: "default", Name: "test"}
				err := processor.ProcessApprovalRequest(ctx, resourceKey, adapter, "approve", req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("approver is required"))
			})

			It("should return error for missing reason", func() {
				alertScale := &opsv1beta1.AlertScale{}
				adapter := handlers.NewAlertScaleApprovalAdapter(alertScale)
				req := handlers.CommonApprovalRequest{
					Approver: "test-approver",
				}

				resourceKey := client.ObjectKey{Namespace: "default", Name: "test"}
				err := processor.ProcessApprovalRequest(ctx, resourceKey, adapter, "approve", req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("reason is required"))
			})
		})
	})
})

var _ = Describe("ApprovalAdapters", func() {
	Describe("AlertScaleApprovalAdapter", func() {
		It("should correctly identify approval state", func() {
			alertScale := &opsv1beta1.AlertScale{
				Status: opsv1beta1.AlertScaleStatus{
					ScaleStatus: opsv1beta1.ScaleStatus{
						Status: types.ScaleStatusApprovaling,
					},
				},
			}
			adapter := handlers.NewAlertScaleApprovalAdapter(alertScale)
			Expect(adapter.IsInApprovalState()).To(BeTrue())

			alertScale.Status.ScaleStatus.Status = types.ScaleStatusApproved
			Expect(adapter.IsInApprovalState()).To(BeFalse())
		})

		It("should return correct status field name", func() {
			alertScale := &opsv1beta1.AlertScale{}
			adapter := handlers.NewAlertScaleApprovalAdapter(alertScale)
			Expect(adapter.GetStatusFieldName()).To(Equal("Status.ScaleStatus.Status"))
		})
	})

	Describe("PodRebalanceApprovalAdapter", func() {
		It("should correctly identify approval state", func() {
			podRebalance := &opsv1beta1.PodRebalance{
				Status: opsv1beta1.PodRebalanceStatus{
					Status: types.RebalanceStatusApprovaling,
				},
			}
			adapter := handlers.NewPodRebalanceApprovalAdapter(podRebalance)
			Expect(adapter.IsInApprovalState()).To(BeTrue())

			podRebalance.Status.Status = types.RebalanceStatusApproved
			Expect(adapter.IsInApprovalState()).To(BeFalse())
		})

		It("should return correct status field name", func() {
			podRebalance := &opsv1beta1.PodRebalance{}
			adapter := handlers.NewPodRebalanceApprovalAdapter(podRebalance)
			Expect(adapter.GetStatusFieldName()).To(Equal("Status.Status"))
		})
	})
})
