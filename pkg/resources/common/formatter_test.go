package common

import (
	"net/url"
	"testing"

	"github.com/acorn-io/brent/pkg/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_includeFields(t *testing.T) {
	tests := []struct {
		name    string
		request *types.APIRequest
		unstr   *unstructured.Unstructured
		want    *unstructured.Unstructured
	}{
		{
			name: "include top level field",
			request: &types.APIRequest{
				Query: url.Values{
					"include": []string{"metadata"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
				},
			},
		},
		{
			name: "include sub field",
			request: &types.APIRequest{
				Query: url.Values{
					"include": []string{"metadata.managedFields"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
						"managedFields": []map[string]interface{}{
							{
								"apiVersion": "v1",
								"fieldsType": "FieldsV1",
								"fieldsV1": map[string]interface{}{
									"f:data": map[string]interface{}{
										".":        map[string]interface{}{},
										"f:ca.crt": map[string]interface{}{},
									},
								},
								"manager":   "kube-controller-manager",
								"operation": "Update",
								"time":      "2022-04-11T22:05:27Z",
							},
						},
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"managedFields": []map[string]interface{}{
							{
								"apiVersion": "v1",
								"fieldsType": "FieldsV1",
								"fieldsV1": map[string]interface{}{
									"f:data": map[string]interface{}{
										".":        map[string]interface{}{},
										"f:ca.crt": map[string]interface{}{},
									},
								},
								"manager":   "kube-controller-manager",
								"operation": "Update",
								"time":      "2022-04-11T22:05:27Z",
							},
						},
					},
				},
			},
		},
		{
			name: "include invalid field",
			request: &types.APIRequest{
				Query: url.Values{
					"include": []string{"foo.bar"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
		},
		{
			name: "include multiple fields",
			request: &types.APIRequest{
				Query: url.Values{
					"include": []string{"kind", "apiVersion", "metadata.name"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "kube-root-ca.crt",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			includeFields(tt.request, tt.unstr)
			assert.Equal(t, tt.want, tt.unstr)
		})
	}
}

func Test_excludeFields(t *testing.T) {
	tests := []struct {
		name    string
		request *types.APIRequest
		unstr   *unstructured.Unstructured
		want    *unstructured.Unstructured
	}{
		{
			name: "exclude top level field",
			request: &types.APIRequest{
				Query: url.Values{
					"exclude": []string{"metadata"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
		},
		{
			name: "exclude sub field",
			request: &types.APIRequest{
				Query: url.Values{
					"exclude": []string{"metadata.managedFields"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
						"managedFields": []map[string]interface{}{
							{
								"apiVersion": "v1",
								"fieldsType": "FieldsV1",
								"fieldsV1": map[string]interface{}{
									"f:data": map[string]interface{}{
										".":        map[string]interface{}{},
										"f:ca.crt": map[string]interface{}{},
									},
								},
								"manager":   "kube-controller-manager",
								"operation": "Update",
								"time":      "2022-04-11T22:05:27Z",
							},
						},
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
		},
		{
			name: "exclude invalid field",
			request: &types.APIRequest{
				Query: url.Values{
					"exclude": []string{"foo.bar"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
		},
		{
			name: "exclude multiple fields",
			request: &types.APIRequest{
				Query: url.Values{
					"exclude": []string{"kind", "apiVersion", "metadata.name"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excludeFields(tt.request, tt.unstr)
			assert.Equal(t, tt.want, tt.unstr)
		})
	}
}

func Test_excludeValues(t *testing.T) {
	tests := []struct {
		name    string
		request *types.APIRequest
		unstr   *unstructured.Unstructured
		want    *unstructured.Unstructured
	}{
		{
			name: "exclude top level value",
			request: &types.APIRequest{
				Query: url.Values{
					"excludeValues": []string{"data"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "",
					},
				},
			},
		},
		{
			name: "exclude sub field value",
			request: &types.APIRequest{
				Query: url.Values{
					"excludeValues": []string{"metadata.annotations"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"deployment.kubernetes.io/revision": "2",
							"meta.helm.sh/release-name":         "fleet-agent-local",
							"meta.helm.sh/release-namespace":    "cattle-fleet-local-system",
						},
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "fleet-agent",
						"namespace":         "cattle-fleet-local-system",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"spec": map[string]interface{}{
						"replicas": 1,
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"deployment.kubernetes.io/revision": "",
							"meta.helm.sh/release-name":         "",
							"meta.helm.sh/release-namespace":    "",
						},
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "fleet-agent",
						"namespace":         "cattle-fleet-local-system",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"spec": map[string]interface{}{
						"replicas": 1,
					},
				},
			},
		},
		{
			name: "exclude invalid value",
			request: &types.APIRequest{
				Query: url.Values{
					"excludeValues": []string{"foo.bar"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "c-m-w466b2vg",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"data": map[string]interface{}{
						"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAg\n-----END CERTIFICATE-----\n",
					},
				},
			},
		},
		{
			name: "exclude multiple values",
			request: &types.APIRequest{
				Query: url.Values{
					"excludeValues": []string{"metadata.annotations", "metadata.labels"},
				},
			},
			unstr: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"deployment.kubernetes.io/revision": "2",
							"meta.helm.sh/release-name":         "fleet-agent-local",
							"meta.helm.sh/release-namespace":    "cattle-fleet-local-system",
						},
						"labels": map[string]interface{}{
							"app.kubernetes.io/managed-by": "Helm",
							"objectset.rio.cattle.io/hash": "362023f752e7f1989d8b652e029bd2c658ae7c44",
						},
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "fleet-agent",
						"namespace":         "cattle-fleet-local-system",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"spec": map[string]interface{}{
						"replicas": 1,
					},
				},
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"deployment.kubernetes.io/revision": "",
							"meta.helm.sh/release-name":         "",
							"meta.helm.sh/release-namespace":    "",
						},
						"labels": map[string]interface{}{
							"app.kubernetes.io/managed-by": "",
							"objectset.rio.cattle.io/hash": "",
						},
						"creationTimestamp": "2022-04-11T22:05:27Z",
						"name":              "fleet-agent",
						"namespace":         "cattle-fleet-local-system",
						"resourceVersion":   "36948",
						"uid":               "1c497934-52cb-42ab-a613-dedfd5fb207b",
					},
					"spec": map[string]interface{}{
						"replicas": 1,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excludeValues(tt.request, tt.unstr)
			assert.Equal(t, tt.want, tt.unstr)
		})
	}
}
