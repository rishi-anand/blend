package main

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v2"
)

var value = `
pack:
  k8sHardening: True
  podCIDR: "192.168.0.0/16"
  serviceClusterIpRange: "10.96.0.0/12"

# KubeAdm customization for kubernetes hardening. Below config will be ignored if k8sHardening property above is disabled
kubeadmconfig:
  apiServer:
    extraArgs:
      # Note : secure-port flag is used during kubeadm init. Do not change this flag on a running cluster
      secure-port: "6443"
      anonymous-auth: "true"
      insecure-port: "0"
      profiling: "false"
      disable-admission-plugins: "AlwaysAdmit"
      default-not-ready-toleration-seconds: "60"
      default-unreachable-toleration-seconds: "60"
      enable-admission-plugins: "AlwaysPullImages,NamespaceLifecycle,ServiceAccount,NodeRestriction,PodSecurityPolicy"
      audit-log-path: /var/log/apiserver/audit.log
      audit-policy-file: /etc/kubernetes/audit-policy.yaml
      audit-log-maxage: "30"
      audit-log-maxbackup: "10"
      audit-log-maxsize: "100"
      authorization-mode: RBAC,Node
      tls-cipher-suites: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256"
    extraVolumes:
      - name: audit-log
        hostPath: /var/log/apiserver
        mountPath: /var/log/apiserver
        pathType: DirectoryOrCreate
      - name: audit-policy
        hostPath: /etc/kubernetes/audit-policy.yaml
        mountPath: /etc/kubernetes/audit-policy.yaml
        readOnly: true
        pathType: File
  controllerManager:
    extraArgs:
      profiling: "false"
      terminated-pod-gc-threshold: "25"
      pod-eviction-timeout: "1m0s"
      use-service-account-credentials: "true"
      feature-gates: "RotateKubeletServerCertificate=true"
  scheduler:
    extraArgs:
      profiling: "false"
  kubeletExtraArgs:
    read-only-port : "0"
    event-qps: "0"
    feature-gates: "RotateKubeletServerCertificate=true"
    protect-kernel-defaults: "true"
    tls-cipher-suites: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256"
  files:
    - path: hardening/audit-policy.yaml
      targetPath: /etc/kubernetes/audit-policy.yaml
      targetOwner: "root:root"
      targetPermissions: "0600"
    - path: hardening/privileged-psp.yaml
      targetPath: /etc/kubernetes/hardening/privileged-psp.yaml
      targetOwner: "root:root"
      targetPermissions: "0600"
    - path: hardening/90-kubelet.conf
      targetPath: /etc/sysctl.d/90-kubelet.conf
      targetOwner: "root:root"
      targetPermissions: "0600"
  preKubeadmCommands:
    # For enabling 'protect-kernel-defaults' flag to kubelet, kernel parameters changes are required
    - 'echo "====> Applying kernel parameters for Kubelet"'
    - 'sysctl -p /etc/sysctl.d/90-kubelet.conf'
  postKubeadmCommands:
    # Apply the privileged PodSecurityPolicy on the first master node ; Otherwise, CNI (and other) pods won't come up
    - 'export KUBECONFIG=/etc/kubernetes/admin.conf'
    # Sometimes api server takes a little longer to respond. Retry if applying the pod-security-policy manifest fails
    - '[ -f "$KUBECONFIG" ] && { echo " ====> Applying PodSecurityPolicy" ; until $(kubectl apply -f /etc/kubernetes/hardening/privileged-psp.yaml > /dev/null ); do echo "Failed to apply PodSecurityPolicies, will retry in 5s" ; sleep 5 ; done ; } || echo "Skipping PodSecurityPolicy for worker nodes"'
`

var override = `
kubeadmconfig:
  apiServer:
    extraArgs:
      secure-port: "6666"
`

var want = `
pack:
  k8sHardening: true
  podCIDR: 192.168.0.0/16
  serviceClusterIpRange: 10.96.0.0/12
kubeadmconfig:
  apiServer:
    extraArgs:
      secure-port: "6666"
      anonymous-auth: "true"
      insecure-port: "0"
      profiling: "false"
      disable-admission-plugins: AlwaysAdmit
      default-not-ready-toleration-seconds: "60"
      default-unreachable-toleration-seconds: "60"
      enable-admission-plugins: AlwaysPullImages,NamespaceLifecycle,ServiceAccount,NodeRestriction,PodSecurityPolicy
      audit-log-path: /var/log/apiserver/audit.log
      audit-policy-file: /etc/kubernetes/audit-policy.yaml
      audit-log-maxage: "30"
      audit-log-maxbackup: "10"
      audit-log-maxsize: "100"
      authorization-mode: RBAC,Node
      tls-cipher-suites: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
    extraVolumes:
    - name: audit-log
      hostPath: /var/log/apiserver
      mountPath: /var/log/apiserver
      pathType: DirectoryOrCreate
    - name: audit-policy
      hostPath: /etc/kubernetes/audit-policy.yaml
      mountPath: /etc/kubernetes/audit-policy.yaml
      readOnly: true
      pathType: File
  controllerManager:
    extraArgs:
      profiling: "false"
      terminated-pod-gc-threshold: "25"
      pod-eviction-timeout: 1m0s
      use-service-account-credentials: "true"
      feature-gates: RotateKubeletServerCertificate=true
  scheduler:
    extraArgs:
      profiling: "false"
  kubeletExtraArgs:
    read-only-port: "0"
    event-qps: "0"
    feature-gates: RotateKubeletServerCertificate=true
    protect-kernel-defaults: "true"
    tls-cipher-suites: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
  files:
  - path: hardening/audit-policy.yaml
    targetPath: /etc/kubernetes/audit-policy.yaml
    targetOwner: root:root
    targetPermissions: "0600"
  - path: hardening/privileged-psp.yaml
    targetPath: /etc/kubernetes/hardening/privileged-psp.yaml
    targetOwner: root:root
    targetPermissions: "0600"
  - path: hardening/90-kubelet.conf
    targetPath: /etc/sysctl.d/90-kubelet.conf
    targetOwner: root:root
    targetPermissions: "0600"
  preKubeadmCommands:
  - echo "====> Applying kernel parameters for Kubelet"
  - sysctl -p /etc/sysctl.d/90-kubelet.conf
  postKubeadmCommands:
  - export KUBECONFIG=/etc/kubernetes/admin.conf
  - '[ -f "$KUBECONFIG" ] && { echo " ====> Applying PodSecurityPolicy" ; until $(kubectl
    apply -f /etc/kubernetes/hardening/privileged-psp.yaml > /dev/null ); do echo
    "Failed to apply PodSecurityPolicies, will retry in 5s" ; sleep 5 ; done ; } ||
    echo "Skipping PodSecurityPolicy for worker nodes"'

`

func TestBlend(t *testing.T) {
	type args struct {
		value    []byte
		override []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"test1", args{value: []byte(value), override: []byte(override)}, []byte(want), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Blend(tt.args.value, tt.args.override)
			if (err != nil) != tt.wantErr {
				t.Errorf("Blend() error = %v, want %v", string(got), string(tt.want))
				return
			}

			//converting got and want to yaml slice for comparison to ignore comments, extra spaces etc
			gotSlice := yaml.MapSlice{}
			yaml.Unmarshal(got, &gotSlice)

			wantSlice := yaml.MapSlice{}
			yaml.Unmarshal(tt.want, &wantSlice)

			byteGot, err := yaml.Marshal(gotSlice)
			if err != nil {
				t.Errorf("Failed to marshal got yaml to byte %v", string(got))
			}

			byteWant, err := yaml.Marshal(wantSlice)
			if err != nil {
				t.Errorf("Failed to marshal want yaml to byte %v", string(tt.want))
			}

			if !bytes.Equal(byteGot, byteWant) {
				t.Errorf("Blend() error = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
