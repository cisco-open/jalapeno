
# Minimal Instance

For local development & testing, you may not have resources for a full installation of Jalapeno. In that case, you might want to spin up a minimal set of Jalapeno's services.

With the minimal deployment:

- No collectors are deployed
- No Kafka services
- Only the [Topology](../about/processors.md#topology-processor) processor is included

To install the minimal version:

1. Clone the Jalapeno repo and `cd` into the folder:

    ```bash
    git clone https://github.com/cisco-open/jalapeno.git && cd jalapeno
    ```

2. Use the `deploy_minimal_jalapeno.sh` script to start the Jalapeno services:

    ```bash
    ./deploy_minimal_jalapeno.sh [path_to_kubectl]
    ```

To destroy the minimal deployment:

1. Use the `destroy_minimal_jalapeno.sh` script:

    ```bash
    destroy_jalapeno.sh kubectl
    ```
