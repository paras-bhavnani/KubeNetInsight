resource "kubernetes_daemonset" "kube_net_insight_ds" {
  metadata {
    name      = "kubenetinsight"
    namespace = "kube-system"
    labels = {
      k8s-app = "kubenetinsight"
    }
  }

  spec {
    selector {
      match_labels = {
        name = "kubenetinsight"
      }
    }

    template {
      metadata {
        labels = {
          name = "kubenetinsight"
        }
      }

      spec {
        toleration {
          key      = "node-role.kubernetes.io/master"
          operator = "Exists"
          effect   = "NoSchedule"
        }

        container {
          name  = "kubenetinsight"
          image = "paras1904/kubenetinsight:latest"

          security_context {
            privileged = true
          }

          volume_mount {
            name       = "bpffs"
            mount_path = "/sys/fs/bpf"
          }

          volume_mount {
            name       = "debugfs"
            mount_path = "/sys/kernel/debug"
          }
        }

        volume {
          name = "bpffs"
          host_path {
            path = "/sys/fs/bpf"
            type = "DirectoryOrCreate"
          }
        }

        volume {
          name = "debugfs"
          host_path {
            path = "/sys/kernel/debug"
          }
        }

        host_network = true
      }
    }
  }
}
