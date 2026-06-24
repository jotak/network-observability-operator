# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:8ff6f941a6c1399cb7d872bb2ff30252d75691589c88ed7949b5f6def6541537'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:3d1d807560d17a73ce2f1c3c4adddd38c221e7f9e9390382dff3e65b96344199'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:f1b2af144675c1ab63ab168fa80e38bd44a001c994dbcf2252f9b956525ddfeb'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:8c4d049b3fed74f51126240d6f8f8220215c67ec887b623673e013ecad593f76'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:d95baa96fa07db96c7987e2d9f9f594c331e995c117af46150462089819ac2f6'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:8a415cd955b8c811c0f641e1a5859a8cbc45ff3a3b268178330107894e9eaafc'
