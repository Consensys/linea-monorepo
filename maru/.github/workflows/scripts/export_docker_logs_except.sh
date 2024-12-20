#!/bin/bash

exception_array=("$@")

mkdir -p docker_logs

is_exception() {
    local container_name="$1"
    for exception in "${exception_array[@]}"; do
        if [[ "$container_name" == "$exception" ]]; then
            return 0
        fi
    done
    return 1
}

containers=$(docker ps --format "{{.ID}} {{.Names}}")

echo "$containers" | while IFS= read -r container_info; do
    container_id=$(echo "$container_info" | awk "{print \$1}")
    container_name=$(echo "$container_info" | awk "{print \$2}")

    if is_exception "$container_name"; then
        echo "Skipping logs for $container_name"
        continue
    fi

    log_file="docker_logs/${container_name}.log"
    docker logs "$container_id" > "$log_file" 2>&1
    echo "Logs exported to $log_file"
done
