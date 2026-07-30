# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:5bd1ceb735e0c6ea0ef6ea0c5452db7d3a4625c7eac579d9ee350cd66b1bee22'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:0db523fcce445d16054aab8e55a186837e76a96ad26aaee015d0611cfd5dc15a'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:74802d90442b055fa2c5c2181179e3903c5b39727b758bc122f785a7a0a7258c'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:a1ecf1a49145120b1d258cfa2c331d99981e5e997fb0413426fbc46c2ee30b88'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:504d825fd7a5513fd7cd49dbaee83a5baf6f1e640eeabd191db9e7d472324a54'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:782cb3b8438386b3c9c35061272bea813e798b60fa6043d79ada6c7e7a5e951e'
