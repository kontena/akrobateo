#!/bin/bash

set -ue

SRC_REPO="docker://quay.io/kontena"
DST_REPO="docker://docker.io/kontenapharos"

create_list() {
    local image=$1
    local version=$2
    local platforms=$3
    local docker_creds=(${DOCKER_CREDS//:/ })
    manifest-tool --username ${docker_creds[0]} --password ${docker_creds[1]} push from-args \
        --platforms $platforms --template "docker.io/kontenapharos/${image}-ARCH:${version}" \
        --target "docker.io/kontenapharos/${image}:${version}"
}

archs=(amd64 arm64)
for ARCH in "${archs[@]}"
do
    image="akrobateo-lb-${ARCH}:${DRONE_TAG}"
    echo "Starting to sync image ${image} ..."
    skopeo --override-arch=${ARCH} copy --dest-creds="${DOCKER_CREDS}" "${SRC_REPO}/${image}" "${DST_REPO}/${image}"

    image="akrobateo-${ARCH}:${DRONE_TAG}"
    skopeo --override-arch=${ARCH} copy --dest-creds="${DOCKER_CREDS}" "${SRC_REPO}/${image}" "${DST_REPO}/${image}"
done


platforms="linux/amd64,linux/arm64"
create_list "akrobateo-lb" "${DRONE_TAG}" $platforms
create_list "akrobateo" "${DRONE_TAG}" $platforms