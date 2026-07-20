#!/usr/bin/env bash

: "${BUILDVERSION:=unknown}"

echo "Updating container file for v${BUILDVERSION}"

# supported_ocp_versions="v4.13"
manifests_dir="./bundles/openshift/manifests"
# metadata_dir="./bundle/metadata"
crd_name="flows.netobserv.io_flowcollectors.yaml"
crd_file="${manifests_dir}/${crd_name}"
csv_name="netobserv-operator.clusterserviceversion.yaml"
csv_file="${manifests_dir}/${csv_name}"

source ./hack/nudging/container_digest.sh

[ ! -f "${crd_file}" ] && { echo "CustomResourceDefinition file not found, the version or name might have changed on us!"; exit 5; }

sed -i 's/\<NetObserv\>/network observability/g' "${crd_file}"

export EPOC_TIMESTAMP=$(date +%s)

VERSION="${BUILDVERSION}" TARGET_CSV_FILE="${csv_file}" python3 ./hack/patch_csv.py

#Using downstream base image
echo "Container file updated"
