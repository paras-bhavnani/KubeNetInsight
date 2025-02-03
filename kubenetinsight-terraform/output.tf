output "daemonset_info" {
  description = "Information about the KubeNetInsight DaemonSet"
  value = {
    name      = kubernetes_daemonset.kube_net_insight_ds.metadata[0].name
    namespace = kubernetes_daemonset.kube_net_insight_ds.metadata[0].namespace
    labels    = kubernetes_daemonset.kube_net_insight_ds.metadata[0].labels
  }
}

output "daemonset_spec" {
  description = "Specification of the KubeNetInsight DaemonSet"
  value = {
    selector = kubernetes_daemonset.kube_net_insight_ds.spec[0].selector
    template = kubernetes_daemonset.kube_net_insight_ds.spec[0].template
  }
}
