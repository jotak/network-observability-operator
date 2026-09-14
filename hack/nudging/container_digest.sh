# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:265edd3ca32da7db9e545e993c2183f961e7ce5981009910637a01ceb1031639'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:c655620d824e66024dadd69236cb786f0eed8df492245a0abf037711fc1432d7'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:7e9a0f36fe3639ac45b6fca03aacc48050943b154df1e834f9eba6308a7d05f5'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:57653ef5bda5330d761fe473c1d14a579f0b34d5dbad091b3b3b54467e831da7'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:12bc71859b5d4d149f7e626740ff71bba241acb4c109c8f15088d8011d70d913'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:f356d9548ef89865dacbc1993875ecd7fb84cb19a2b4a0d5ea0ca549b7829b4a'
