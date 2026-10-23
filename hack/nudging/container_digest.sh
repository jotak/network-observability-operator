# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:8e8001401041922296e27790347674fa1d993445ddc84f37f20d969b527c514a'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:76499a26371f966e02d30a2512d116a049faf939d7f0106e61b3c4d1a221be7b'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:e8867578dc060675325e81a25a5a3ca1e06c7e7a43759269f3b8a221c3efd147'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:c122818a8b15f8381d3f3f4f81f33a16d8c275db54389943cecacfd3fc96246f'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:ee767720a1d79b5cde5c11ec99c952a34a85c55f46d25ba78eb453c9c2c22c03'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:4bb5904583b2a739d103fcad761d34ec968e563282b8f3eb675f90e0d351779e'
